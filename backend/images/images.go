// Service images gestisce il catalogo e lo storage delle immagini.
package images

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

var db = sqldb.NewDatabase("images", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// imagesBucket stores image files in Encore object storage.
var imagesBucket = objects.NewBucket("images", objects.BucketConfig{
	Public: true,
})

// ImageCategory categorizza le immagini caricate.
type ImageCategory string

const (
	ImageCategoryGeneral ImageCategory = "general"
	ImageCategoryAvatar  ImageCategory = "avatar"
	ImageCategoryPromo   ImageCategory = "promo"
	ImageCategoryBanner  ImageCategory = "banner"
)

// Image rappresenta una risorsa immagine salvata nel database e nello storage.
type Image struct {
	UUID        uuid.UUID `json:"uuid"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	URL         string    `json:"url"`
	Proportion  string    `json:"proportion"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
}

type ListImagesResponse struct {
	Images []*Image `json:"images"`
}

type SaveImageParams struct {
	UUID        string `json:"uuid,omitempty"`
	Description string `json:"description"`
	Data        string `json:"data,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Proportion  string `json:"proportion,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

func (p *SaveImageParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "i parametri immagine sono obbligatori"}
	}
	if strings.TrimSpace(p.Description) == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "la descrizione dell'immagine è obbligatoria"}
	}
	if strings.TrimSpace(p.UUID) == "" && strings.TrimSpace(p.Data) == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "il contenuto dell'immagine (data) è obbligatorio per una nuova immagine"}
	}
	return nil
}

func parseImageData(data string, fallbackContentType string) ([]byte, string, error) {
	data = strings.TrimSpace(data)
	contentType := fallbackContentType
	if contentType == "" {
		contentType = "image/png"
	}

	if strings.HasPrefix(data, "data:") {
		parts := strings.SplitN(data, ",", 2)
		if len(parts) == 2 {
			header := parts[0]
			data = parts[1]
			if strings.Contains(header, ";") {
				ct := strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
				if ct != "" {
					contentType = ct
				}
			}
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", errs.WrapCode(err, errs.InvalidArgument, "dati immagine base64 non validi")
	}

	return decoded, contentType, nil
}

// ListImages ritorna la lista completa delle immagini.
//
//encore:api public method=GET path=/images
func ListImages(ctx context.Context) (*ListImagesResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT uuid, description, created_at
		FROM images
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante il recupero delle immagini")
	}
	defer rows.Close()

	var list []*Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.UUID, &img.Description, &img.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del record immagine")
		}
		img.URL = imagesBucket.PublicURL(img.UUID.String()).String()
		list = append(list, &img)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante l'iterazione delle immagini")
	}

	if list == nil {
		list = []*Image{}
	}

	return &ListImagesResponse{Images: list}, nil
}

// GetImage ritorna lo stream binario dell'immagine memorizzata nello store.
//
//encore:api public raw method=GET path=/image/:uuid
func GetImage(w http.ResponseWriter, req *http.Request) {
	uuidStr := req.PathValue("uuid")
	if uuidStr == "" {
		uuidStr = strings.TrimPrefix(req.URL.Path, "/image/")
	}
	uuidStr = strings.TrimSpace(uuidStr)

	parsedUUID, err := uuid.FromString(uuidStr)
	if err != nil {
		errs.HTTPError(w, &errs.Error{Code: errs.InvalidArgument, Message: "UUID immagine non valido"})
		return
	}

	reader := imagesBucket.Download(req.Context(), parsedUUID.String())
	defer reader.Close()

	w.Header().Set("Content-Type", "image/png")
	if _, err := io.Copy(w, reader); err != nil {
		if reader.Err() != nil {
			errs.HTTPError(w, &errs.Error{Code: errs.NotFound, Message: "immagine non trovata"})
			return
		}
	}
	if reader.Err() != nil {
		errs.HTTPError(w, &errs.Error{Code: errs.NotFound, Message: "immagine non trovata"})
		return
	}
}

// SaveImage crea una nuova immagine o aggiorna la descrizione/file dell'immagine esistente.
//
//encore:api public method=POST path=/image
func SaveImage(ctx context.Context, p *SaveImageParams) (*Image, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var targetUUID uuid.UUID
	var isUpdate bool

	uuidStr := strings.TrimSpace(p.UUID)
	if uuidStr != "" {
		parsedUUID, err := uuid.FromString(uuidStr)
		if err != nil {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "UUID immagine non valido"}
		}
		targetUUID = parsedUUID
		var exists bool
		err = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM images WHERE uuid = $1)`, targetUUID).Scan(&exists)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante il controllo esistenza immagine")
		}
		isUpdate = exists
	} else {
		newID, err := uuid.NewV4()
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile generare l'identificativo UUID")
		}
		targetUUID = newID
	}

	if strings.TrimSpace(p.Data) != "" {
		rawBytes, detectedContentType, err := parseImageData(p.Data, p.ContentType)
		if err != nil {
			return nil, err
		}

		writer := imagesBucket.Upload(ctx, targetUUID.String(), objects.WithUploadAttrs(objects.UploadAttrs{
			ContentType: detectedContentType,
		}))
		if _, err := io.Copy(writer, bytes.NewReader(rawBytes)); err != nil {
			writer.Abort(err)
			return nil, errs.WrapCode(err, errs.Internal, "errore durante il salvataggio dell'immagine nello store")
		}
		if err := writer.Close(); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la chiusura dello store immagine")
		}
	}

	var createdAt time.Time
	if isUpdate {
		err := db.QueryRow(ctx, `
			UPDATE images
			SET description = $1
			WHERE uuid = $2
			RETURNING created_at
		`, strings.TrimSpace(p.Description), targetUUID).Scan(&createdAt)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare l'immagine nel database")
		}
	} else {
		err := db.QueryRow(ctx, `
			INSERT INTO images (uuid, description)
			VALUES ($1, $2)
			RETURNING created_at
		`, targetUUID, strings.TrimSpace(p.Description)).Scan(&createdAt)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile creare il record immagine nel database")
		}
	}

	return &Image{
		UUID:        targetUUID,
		Description: strings.TrimSpace(p.Description),
		CreatedAt:   createdAt,
		URL:         imagesBucket.PublicURL(targetUUID.String()).String(),
		Proportion:  p.Proportion,
		Width:       p.Width,
		Height:      p.Height,
	}, nil
}

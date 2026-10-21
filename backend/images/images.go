// Service images gestisce il catalogo e lo storage delle immagini.
package images

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	authsvc "encore.app/auth"
	"encore.app/validate"
	"encore.dev/beta/auth"
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

type ImageDescription struct {
	Text  string `json:"text"`
	Box_X int    `json:"box_x"`
	Box_y int    `json:"box_y"`
}

// UnmarshalJSON supports both plain text string ("description text") and structured object ({ "text": "...", "box_x": ..., "box_y": ... }).
func (d *ImageDescription) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*d = ImageDescription{
			Text: text,
		}
		return nil
	}

	type Alias ImageDescription
	var aux struct {
		Alias
		AltBoxY  int `json:"box_Y"`
		AltBoxX  int `json:"box_X"`
		CapBoxX  int `json:"Box_X"`
		CapBoxY  int `json:"Box_y"`
		CapBoxY2 int `json:"Box_Y"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*d = ImageDescription(aux.Alias)
	if d.Box_X == 0 {
		if aux.AltBoxX != 0 {
			d.Box_X = aux.AltBoxX
		} else if aux.CapBoxX != 0 {
			d.Box_X = aux.CapBoxX
		}
	}
	if d.Box_y == 0 {
		if aux.AltBoxY != 0 {
			d.Box_y = aux.AltBoxY
		} else if aux.CapBoxY != 0 {
			d.Box_y = aux.CapBoxY
		} else if aux.CapBoxY2 != 0 {
			d.Box_y = aux.CapBoxY2
		}
	}
	return nil
}

// Value implements driver.Valuer for database storage.
func (d ImageDescription) Value() (driver.Value, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner for database retrieval.
func (d *ImageDescription) Scan(src any) error {
	if src == nil {
		*d = ImageDescription{}
		return nil
	}
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for ImageDescription: %T", src)
	}

	str = strings.TrimSpace(str)
	if str == "" {
		*d = ImageDescription{}
		return nil
	}

	if strings.HasPrefix(str, "{") {
		if err := json.Unmarshal([]byte(str), d); err == nil {
			return nil
		}
	}

	*d = ImageDescription{
		Text: str,
	}
	return nil
}

// Image rappresenta una risorsa immagine salvata nel database e nello storage.
type Image struct {
	UUID        uuid.UUID        `json:"uuid"`
	Owner       uuid.UUID        `json:"owner"`
	Description ImageDescription `json:"description"`
	Approved    bool             `json:"approved"`
	CreatedAt   time.Time        `json:"created_at"`
	URL         string           `json:"url"`
	Proportion  string           `json:"proportion"`
	Width       int              `json:"width"`
	Height      int              `json:"height"`
}

type ListImagesResponse struct {
	Images []*Image `json:"images"`
}

type SaveImageParams struct {
	UUID        string           `json:"uuid,omitempty"`
	Description ImageDescription `json:"description"`
	Data        string           `json:"data,omitempty"`
	ContentType string           `json:"content_type,omitempty"`
	Proportion  string           `json:"proportion,omitempty"`
	Width       int              `json:"width,omitempty"`
	Height      int              `json:"height,omitempty"`
}

func (p *SaveImageParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "i parametri immagine sono obbligatori"}
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

// ListImages ritorna le immagini visibili al chiamante: agli anonimi solo
// quelle approvate, agli utenti anche le proprie, agli admin tutte.
//
//encore:api public method=GET path=/images
func ListImages(ctx context.Context) (*ListImagesResponse, error) {
	all := authsvc.IsAdmin()
	owner := uuid.Nil
	if uid, ok := auth.UserID(); ok {
		owner, _ = uuid.FromString(string(uid))
	}
	rows, err := db.Query(ctx, `
		SELECT uuid, owner, description, approved, proportion, width, height, created_at
		FROM images
		WHERE $1 OR approved OR ($2 <> '00000000-0000-0000-0000-000000000000'::uuid AND owner = $2)
		ORDER BY created_at DESC
	`, all, owner)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante il recupero delle immagini")
	}
	defer rows.Close()

	var list []*Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(
			&img.UUID,
			&img.Owner,
			&img.Description,
			&img.Approved,
			&img.Proportion,
			&img.Width,
			&img.Height,
			&img.CreatedAt,
		); err != nil {
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
	w.Header().Set("X-Content-Type-Options", "nosniff")
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
// Una nuova immagine appartiene a chi la carica; un'immagine esistente può
// essere modificata solo dal proprietario o da un admin.
//
//encore:api auth method=POST path=/image
func SaveImage(ctx context.Context, p *SaveImageParams) (*Image, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "autenticazione richiesta"}
	}
	callerID, err := uuid.FromString(string(uid))
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "identificativo utente non valido"}
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
		var currentOwner uuid.UUID
		err = db.QueryRow(ctx, `SELECT owner FROM images WHERE uuid = $1`, targetUUID).Scan(&currentOwner)
		switch {
		case errors.Is(err, sqldb.ErrNoRows):
			// Il client non può scegliere l'UUID di una nuova immagine.
			return nil, &errs.Error{Code: errs.NotFound, Message: "immagine non trovata"}
		case err != nil:
			return nil, errs.WrapCode(err, errs.Internal, "errore durante il controllo esistenza immagine")
		case currentOwner != callerID && !authsvc.IsAdmin():
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non sei il proprietario dell'immagine"}
		}
		isUpdate = true
	} else {
		newID, err := uuid.NewV4()
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile generare l'identificativo UUID")
		}
		targetUUID = newID
	}

	if strings.TrimSpace(p.Data) != "" {
		rawBytes, _, err := parseImageData(p.Data, p.ContentType)
		if err != nil {
			return nil, err
		}
		// Il content type dichiarato (data: URL o campo) viene ignorato:
		// conta solo quello rilevato dai byte.
		detectedContentType, err := validate.ImageContentType(rawBytes)
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

	desc := p.Description
	desc.Text = validate.SanitizeInlineHTML(strings.TrimSpace(desc.Text))

	var (
		createdAt time.Time
		resOwner  uuid.UUID
		resAppr   bool
		resProp   string
		resWidth  int
		resHeight int
	)

	if isUpdate {
		err := db.QueryRow(ctx, `
			UPDATE images
			SET description = $1,
			    proportion = CASE WHEN $2 = '' THEN proportion ELSE $2 END,
			    width = CASE WHEN $3 = 0 THEN width ELSE $3 END,
			    height = CASE WHEN $4 = 0 THEN height ELSE $4 END
			WHERE uuid = $5
			RETURNING created_at, owner, approved, proportion, width, height
		`, desc, p.Proportion, p.Width, p.Height, targetUUID).Scan(
			&createdAt, &resOwner, &resAppr, &resProp, &resWidth, &resHeight,
		)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare l'immagine nel database")
		}
	} else {
		err := db.QueryRow(ctx, `
			INSERT INTO images (uuid, owner, description, proportion, width, height)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING created_at, owner, approved, proportion, width, height
		`, targetUUID, callerID, desc, p.Proportion, p.Width, p.Height).Scan(
			&createdAt, &resOwner, &resAppr, &resProp, &resWidth, &resHeight,
		)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "impossibile creare il record immagine nel database")
		}
	}

	return &Image{
		UUID:        targetUUID,
		Owner:       resOwner,
		Description: desc,
		Approved:    resAppr,
		CreatedAt:   createdAt,
		URL:         imagesBucket.PublicURL(targetUUID.String()).String(),
		Proportion:  resProp,
		Width:       resWidth,
		Height:      resHeight,
	}, nil
}

type SetDescriptionParams struct {
	Description ImageDescription `json:"description"`
}

// SetDescription aggiorna solo la descrizione di un'immagine esistente.
// Privato: usato da broadcasters per allineare le didascalie delle foto, dopo
// aver verificato che il chiamante sia proprietario del profilo broadcaster.
//
//encore:api private method=PUT path=/images/internal/:id/description
func SetDescription(ctx context.Context, id uuid.UUID, p *SetDescriptionParams) error {
	desc := p.Description
	desc.Text = validate.SanitizeInlineHTML(strings.TrimSpace(desc.Text))
	if _, err := db.Exec(ctx, `UPDATE images SET description = $1 WHERE uuid = $2`, desc, id); err != nil {
		return errs.WrapCode(err, errs.Internal, "impossibile aggiornare la descrizione dell'immagine")
	}
	return nil
}

// ImageSize sono le dimensioni salvate di un'immagine (0 / "" se non note).
type ImageSize struct {
	UUID       uuid.UUID `json:"uuid"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Proportion string    `json:"proportion"`
}

type GetImageSizesParams struct {
	UUIDs []uuid.UUID `json:"uuids"`
}

type GetImageSizesResponse struct {
	Images []*ImageSize `json:"images"`
}

// GetImageSizes restituisce le dimensioni delle immagini indicate; quelle
// inesistenti vengono omesse. Privato: usato da broadcasters per le promo.
//
//encore:api private method=POST path=/images/internal/sizes
func GetImageSizes(ctx context.Context, p *GetImageSizesParams) (*GetImageSizesResponse, error) {
	resp := &GetImageSizesResponse{Images: []*ImageSize{}}
	if p == nil || len(p.UUIDs) == 0 {
		return resp, nil
	}
	rows, err := db.Query(ctx, `
		SELECT uuid, width, height, proportion FROM images WHERE uuid = ANY($1)
	`, p.UUIDs)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile recuperare le dimensioni delle immagini")
	}
	defer rows.Close()
	for rows.Next() {
		var s ImageSize
		if err := rows.Scan(&s.UUID, &s.Width, &s.Height, &s.Proportion); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura delle dimensioni")
		}
		resp.Images = append(resp.Images, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante l'iterazione delle immagini")
	}
	return resp, nil
}

// Service broadcasters provides APIs and storage for broadcaster profiles and metadata.
package broadcasters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authsvc "encore.app/auth"
	"encore.app/config"
	images "encore.app/images"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

var db = sqldb.NewDatabase("broadcasters", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type StatusResponse struct {
	Status string `json:"status"`
}

// Status verifica lo stato di disponibilità del servizio broadcasters.
//
//encore:api public method=GET path=/broadcasters/status
func Status(ctx context.Context) (*StatusResponse, error) {
	return &StatusResponse{Status: "ok"}, nil
}

type BroadcasterImage struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Proportion string `json:"proportion"`
}

var BroadcasterImageSize = BroadcasterImage{
	Width:      600,
	Height:     0,
	Proportion: "2:3",
}

type BroadcasterPhoto struct {
	ImageURL    string                  `json:"image_url"`
	Alt         string                  `json:"alt"`
	Box_X       int                     `json:"box_x"`
	Box_y       int                     `json:"box_y"`
	Description images.ImageDescription `json:"description"`
}

// UnmarshalJSON supports case variants for Box_x and Box_y and populates Description.
func (bp *BroadcasterPhoto) UnmarshalJSON(data []byte) error {
	type Alias BroadcasterPhoto
	var aux struct {
		Alias
		AltBoxX        int             `json:"box_X"`
		CapBoxX        int             `json:"Box_X"`
		AltBoxY        int             `json:"box_Y"`
		CapBoxY        int             `json:"Box_y"`
		CapBoxY2       int             `json:"Box_Y"`
		RawDescription json.RawMessage `json:"description"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*bp = BroadcasterPhoto(aux.Alias)
	if bp.Box_X == 0 {
		if aux.AltBoxX != 0 {
			bp.Box_X = aux.AltBoxX
		} else if aux.CapBoxX != 0 {
			bp.Box_X = aux.CapBoxX
		}
	}
	if bp.Box_y == 0 {
		if aux.AltBoxY != 0 {
			bp.Box_y = aux.AltBoxY
		} else if aux.CapBoxY != 0 {
			bp.Box_y = aux.CapBoxY
		} else if aux.CapBoxY2 != 0 {
			bp.Box_y = aux.CapBoxY2
		}
	}

	if len(aux.RawDescription) > 0 {
		var imgDesc images.ImageDescription
		if err := json.Unmarshal(aux.RawDescription, &imgDesc); err == nil {
			bp.Description = imgDesc
		} else {
			var plainDesc string
			if err := json.Unmarshal(aux.RawDescription, &plainDesc); err == nil {
				bp.Description.Text = plainDesc
			}
		}
	}

	if bp.Description.Text != "" && bp.Alt == "" {
		bp.Alt = bp.Description.Text
	} else if bp.Alt != "" && bp.Description.Text == "" {
		bp.Description.Text = bp.Alt
	}

	if bp.Description.Box_X != 0 && bp.Box_X == 0 {
		bp.Box_X = bp.Description.Box_X
	} else if bp.Box_X != 0 && bp.Description.Box_X == 0 {
		bp.Description.Box_X = bp.Box_X
	}

	if bp.Description.Box_y != 0 && bp.Box_y == 0 {
		bp.Box_y = bp.Description.Box_y
	} else if bp.Box_y != 0 && bp.Description.Box_y == 0 {
		bp.Description.Box_y = bp.Box_y
	}

	return nil
}

// MarshalJSON ensures both alt/box_x/box_y and description are always serialized.
func (bp BroadcasterPhoto) MarshalJSON() ([]byte, error) {
	desc := bp.Description
	if desc.Text == "" && bp.Alt != "" {
		desc.Text = bp.Alt
	}
	if desc.Box_X == 0 && bp.Box_X != 0 {
		desc.Box_X = bp.Box_X
	}
	if desc.Box_y == 0 && bp.Box_y != 0 {
		desc.Box_y = bp.Box_y
	}

	alt := bp.Alt
	if alt == "" && desc.Text != "" {
		alt = desc.Text
	}
	boxX := bp.Box_X
	if boxX == 0 && desc.Box_X != 0 {
		boxX = desc.Box_X
	}
	boxY := bp.Box_y
	if boxY == 0 && desc.Box_y != 0 {
		boxY = desc.Box_y
	}

	type Alias BroadcasterPhoto
	return json.Marshal(&struct {
		Alias
		Alt         string                  `json:"alt"`
		Box_X       int                     `json:"box_x"`
		Box_y       int                     `json:"box_y"`
		Description images.ImageDescription `json:"description"`
	}{
		Alias:       Alias(bp),
		Alt:         alt,
		Box_X:       boxX,
		Box_y:       boxY,
		Description: desc,
	})
}

func alignPhotos(photos []BroadcasterPhoto) {
	for i := range photos {
		if photos[i].Description.Text != "" && photos[i].Alt == "" {
			photos[i].Alt = photos[i].Description.Text
		} else if photos[i].Alt != "" && photos[i].Description.Text == "" {
			photos[i].Description.Text = photos[i].Alt
		}
		if photos[i].Description.Box_X != 0 && photos[i].Box_X == 0 {
			photos[i].Box_X = photos[i].Description.Box_X
		} else if photos[i].Box_X != 0 && photos[i].Description.Box_X == 0 {
			photos[i].Description.Box_X = photos[i].Box_X
		}
		if photos[i].Description.Box_y != 0 && photos[i].Box_y == 0 {
			photos[i].Box_y = photos[i].Description.Box_y
		} else if photos[i].Box_y != 0 && photos[i].Description.Box_y == 0 {
			photos[i].Description.Box_y = photos[i].Box_y
		}
	}
}

type BroadcasterTimeSlot struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type BroadcasterHours struct {
	Day   string                `json:"day"`
	Hours []BroadcasterTimeSlot `json:"hours"`
}

func (bh *BroadcasterHours) UnmarshalJSON(data []byte) error {
	type rawHours struct {
		Day   string          `json:"day"`
		Hours json.RawMessage `json:"hours"`
	}
	var raw rawHours
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	bh.Day = raw.Day
	bh.Hours = []BroadcasterTimeSlot{}
	if len(raw.Hours) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw.Hours, &bh.Hours); err == nil {
		return nil
	}
	var str string
	if err := json.Unmarshal(raw.Hours, &str); err == nil && str != "" {
		parts := strings.Split(str, "-")
		if len(parts) == 2 {
			bh.Hours = append(bh.Hours, BroadcasterTimeSlot{
				From: strings.TrimSpace(parts[0]),
				To:   strings.TrimSpace(parts[1]),
			})
		}
	}
	return nil
}

func ValidateHours(hours []BroadcasterHours) error {
	for _, h := range hours {
		day := strings.TrimSpace(h.Day)
		if day == "" {
			return &errs.Error{Code: errs.InvalidArgument, Message: "il giorno non può essere vuoto"}
		}
		for i := 0; i < len(h.Hours); i++ {
			slotA := h.Hours[i]
			fromA := strings.TrimSpace(slotA.From)
			toA := strings.TrimSpace(slotA.To)
			if fromA == "" || toA == "" {
				return &errs.Error{Code: errs.InvalidArgument, Message: fmt.Sprintf("orario incompleto per %s: 'from' e 'to' sono obbligatori", day)}
			}
			if fromA >= toA {
				return &errs.Error{Code: errs.InvalidArgument, Message: fmt.Sprintf("orario non valido per %s: l'orario di inizio (%s) deve essere precedente a quello di fine (%s)", day, fromA, toA)}
			}
			for j := i + 1; j < len(h.Hours); j++ {
				slotB := h.Hours[j]
				fromB := strings.TrimSpace(slotB.From)
				toB := strings.TrimSpace(slotB.To)
				if fromA < toB && fromB < toA {
					return &errs.Error{
						Code:    errs.InvalidArgument,
						Message: fmt.Sprintf("orari sovrapposti per %s: la fascia %s-%s si sovrappone con %s-%s", day, fromA, toA, fromB, toB),
					}
				}
			}
		}
	}
	return nil
}

type BroadcasterAttribute struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Broadcaster struct {
	ID            uuid.UUID              `json:"id"`
	UserID        uuid.UUID              `json:"user_id"`
	DisplayName   string                 `json:"display_name"`
	Category      string                 `json:"category"`
	CategoryColor string                 `json:"category_color"`
	Bio           string                 `json:"bio"`
	Certified     bool                   `json:"certified"`
	Availability  string                 `json:"availability"`
	Phone         string                 `json:"phone"`
	Services      []string               `json:"services"`
	Languages     []string               `json:"languages"`
	Attributes    []BroadcasterAttribute `json:"attributes"`
	Hours         []BroadcasterHours     `json:"hours"`
	Photos        []BroadcasterPhoto     `json:"photos"`
	Enabled       bool                   `json:"enabled"`
	Used          bool                   `json:"used"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type ListBroadcastersParams struct {
	Category     string `json:"category,omitempty" query:"category"`
	Certified    bool   `json:"certified,omitempty" query:"certified"`
	Availability string `json:"availability,omitempty" query:"availability"`
	Search       string `json:"search,omitempty" query:"search"`
}

type ListBroadcastersResponse struct {
	Broadcasters []*Broadcaster `json:"broadcasters"`
	Total        int            `json:"total"`
}

type CreateBroadcasterParams struct {
	UserID        string                 `json:"user_id,omitempty"`
	DisplayName   string                 `json:"display_name"`
	Category      string                 `json:"category"`
	CategoryColor string                 `json:"category_color"`
	Bio           string                 `json:"bio"`
	Certified     bool                   `json:"certified"`
	Availability  string                 `json:"availability"`
	Phone         string                 `json:"phone"`
	Services      []string               `json:"services"`
	Languages     []string               `json:"languages"`
	Attributes    []BroadcasterAttribute `json:"attributes"`
	Hours         []BroadcasterHours     `json:"hours"`
	Photos        []BroadcasterPhoto     `json:"photos"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	Used          *bool                  `json:"used,omitempty"`
}

func (p *CreateBroadcasterParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "parametri broadcaster richiesti"}
	}
	p.DisplayName = strings.TrimSpace(p.DisplayName)
	if p.DisplayName == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "display_name è obbligatorio"}
	}
	p.Category = strings.TrimSpace(p.Category)
	if p.Category == "" {
		if len(config.SiteConfig.Categories) > 0 {
			p.Category = config.SiteConfig.Categories[0].Name
		} else {
			p.Category = "General"
		}
	}
	p.CategoryColor = config.GetCategoryColor(p.Category)
	if p.Availability == "" {
		if opts := config.GetAvailabilityOptions(); len(opts) > 0 {
			p.Availability = opts[0]
		} else {
			p.Availability = "available"
		}
	}
	if p.Services == nil {
		p.Services = []string{}
	}
	if p.Languages == nil {
		p.Languages = []string{}
	}
	if p.Attributes == nil {
		p.Attributes = []BroadcasterAttribute{}
	}
	if p.Hours == nil {
		p.Hours = []BroadcasterHours{}
	} else {
		if err := ValidateHours(p.Hours); err != nil {
			return err
		}
	}
	if p.Photos == nil {
		p.Photos = []BroadcasterPhoto{}
	}
	return nil
}

type UpdateBroadcasterParams struct {
	DisplayName   string                 `json:"display_name"`
	Category      string                 `json:"category"`
	CategoryColor string                 `json:"category_color"`
	Bio           string                 `json:"bio"`
	Certified     bool                   `json:"certified"`
	Availability  string                 `json:"availability"`
	Phone         string                 `json:"phone"`
	Services      []string               `json:"services"`
	Languages     []string               `json:"languages"`
	Attributes    []BroadcasterAttribute `json:"attributes"`
	Hours         []BroadcasterHours     `json:"hours"`
	Photos        []BroadcasterPhoto     `json:"photos"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	Used          *bool                  `json:"used,omitempty"`
}

func (p *UpdateBroadcasterParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "parametri aggiornamento richiesti"}
	}
	p.DisplayName = strings.TrimSpace(p.DisplayName)
	if p.DisplayName == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "display_name non può essere vuoto"}
	}
	p.Category = strings.TrimSpace(p.Category)
	if p.Category == "" {
		if len(config.SiteConfig.Categories) > 0 {
			p.Category = config.SiteConfig.Categories[0].Name
		} else {
			p.Category = "General"
		}
	}
	p.CategoryColor = config.GetCategoryColor(p.Category)
	if p.Availability == "" {
		if opts := config.GetAvailabilityOptions(); len(opts) > 0 {
			p.Availability = opts[0]
		} else {
			p.Availability = "available"
		}
	}
	if p.Services == nil {
		p.Services = []string{}
	}
	if p.Languages == nil {
		p.Languages = []string{}
	}
	if p.Hours != nil {
		if err := ValidateHours(p.Hours); err != nil {
			return err
		}
	}
	return nil
}

func authedUser() (uuid.UUID, bool, error) {
	uid, ok := auth.UserID()
	if !ok {
		return uuid.Nil, false, &errs.Error{Code: errs.Unauthenticated, Message: "autenticazione richiesta"}
	}
	parsedID, err := uuid.FromString(string(uid))
	if err != nil {
		return uuid.Nil, false, &errs.Error{Code: errs.Unauthenticated, Message: "identificativo utente non valido"}
	}
	data, _ := auth.Data().(*authsvc.AuthData)
	isAdmin := data != nil && data.Role.IsAdmin()
	return parsedID, isAdmin, nil
}

func scanBroadcaster(scanner interface{ Scan(dest ...any) error }) (*Broadcaster, error) {
	var (
		b               Broadcaster
		servicesBytes   []byte
		languagesBytes  []byte
		attributesBytes []byte
		hoursBytes      []byte
		photosBytes     []byte
	)

	err := scanner.Scan(
		&b.ID,
		&b.UserID,
		&b.DisplayName,
		&b.Category,
		&b.CategoryColor,
		&b.Bio,
		&b.Certified,
		&b.Availability,
		&b.Phone,
		&servicesBytes,
		&languagesBytes,
		&attributesBytes,
		&hoursBytes,
		&photosBytes,
		&b.Enabled,
		&b.Used,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.Services = []string{}
	if len(servicesBytes) > 0 {
		_ = json.Unmarshal(servicesBytes, &b.Services)
	}

	b.Languages = []string{}
	if len(languagesBytes) > 0 {
		_ = json.Unmarshal(languagesBytes, &b.Languages)
	}

	b.Attributes = []BroadcasterAttribute{}
	if len(attributesBytes) > 0 {
		_ = json.Unmarshal(attributesBytes, &b.Attributes)
	}

	b.Hours = []BroadcasterHours{}
	if len(hoursBytes) > 0 {
		_ = json.Unmarshal(hoursBytes, &b.Hours)
	}

	b.Photos = []BroadcasterPhoto{}
	if len(photosBytes) > 0 {
		_ = json.Unmarshal(photosBytes, &b.Photos)
	}

	return &b, nil
}

// ListBroadcasters restituisce l'elenco dei profili broadcaster con filtri opzionali.
//
//encore:api public method=GET path=/broadcasters
func ListBroadcasters(ctx context.Context, p *ListBroadcastersParams) (*ListBroadcastersResponse, error) {
	var (
		whereClauses []string
		args         []any
		argIdx       = 1
	)

	if p != nil {
		if cat := strings.TrimSpace(p.Category); cat != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIdx))
			args = append(args, cat)
			argIdx++
		}
		if p.Certified {
			whereClauses = append(whereClauses, fmt.Sprintf("certified = $%d", argIdx))
			args = append(args, p.Certified)
			argIdx++
		}
		if avail := strings.TrimSpace(p.Availability); avail != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("availability = $%d", argIdx))
			args = append(args, avail)
			argIdx++
		}
		if q := strings.TrimSpace(p.Search); q != "" {
			like := "%" + q + "%"
			whereClauses = append(whereClauses, fmt.Sprintf("(display_name ILIKE $%d OR bio ILIKE $%d)", argIdx, argIdx))
			args = append(args, like)
			argIdx++
		}
	}

	query := `
		SELECT id, user_id, display_name, category, category_color, bio, certified,
		       availability, phone, services, languages, attributes, hours, photos,
		       enabled, used, created_at, updated_at
		FROM broadcasters
	`
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY certified DESC, display_name ASC, id DESC"

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile recuperare i broadcaster")
	}
	defer rows.Close()

	var result []*Broadcaster
	for rows.Next() {
		b, err := scanBroadcaster(rows)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del record broadcaster")
		}
		result = append(result, b)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante l'iterazione dei broadcaster")
	}

	if result == nil {
		result = []*Broadcaster{}
	}

	return &ListBroadcastersResponse{
		Broadcasters: result,
		Total:        len(result),
	}, nil
}

// GetBroadcaster restituisce il profilo broadcaster dato il suo ID UUID.
//
//encore:api public method=GET path=/broadcasters/id/:id
func GetBroadcaster(ctx context.Context, id uuid.UUID) (*Broadcaster, error) {
	row := db.QueryRow(ctx, `
		SELECT id, user_id, display_name, category, category_color, bio, certified,
		       availability, phone, services, languages, attributes, hours, photos,
		       enabled, used, created_at, updated_at
		FROM broadcasters
		WHERE id = $1
	`, id)

	b, err := scanBroadcaster(row)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster non trovato"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del broadcaster")
	}
	return b, nil
}

// GetBroadcasterByUserID restituisce il profilo broadcaster associato all'UUID utente.
//
//encore:api public method=GET path=/broadcasters/user/:userID
func GetBroadcasterByUserID(ctx context.Context, userID uuid.UUID) (*Broadcaster, error) {
	row := db.QueryRow(ctx, `
		SELECT id, user_id, display_name, category, category_color, bio, certified,
		       availability, phone, services, languages, attributes, hours, photos,
		       enabled, used, created_at, updated_at
		FROM broadcasters
		WHERE user_id = $1
	`, userID)

	b, err := scanBroadcaster(row)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster non trovato per l'utente specificato"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del broadcaster")
	}
	return b, nil
}

// CreateBroadcaster registra un nuovo profilo broadcaster (richiede autenticazione).
//
//encore:api auth method=POST path=/broadcasters
func CreateBroadcaster(ctx context.Context, p *CreateBroadcasterParams) (*Broadcaster, error) {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	targetUserID := callerID
	if uidStr := strings.TrimSpace(p.UserID); uidStr != "" {
		parsedUID, err := uuid.FromString(uidStr)
		if err != nil {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "user_id non valido"}
		}
		if !isAdmin && parsedUID != callerID {
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: solo gli amministratori possono registrare altri broadcaster"}
		}
		targetUserID = parsedUID
	}

	alignPhotos(p.Photos)
	servicesJSON, _ := json.Marshal(p.Services)
	languagesJSON, _ := json.Marshal(p.Languages)
	attributesJSON, _ := json.Marshal(p.Attributes)
	hoursJSON, _ := json.Marshal(p.Hours)
	photosJSON, _ := json.Marshal(p.Photos)

	now := time.Now()
	var existingID uuid.UUID
	checkErr := db.QueryRow(ctx, `SELECT id FROM broadcasters WHERE user_id = $1`, targetUserID).Scan(&existingID)
	if checkErr == nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("esiste già un profilo broadcaster associato a questo utente (ID %s)", existingID),
		}
	}

	broadcasterID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile generare id broadcaster")
	}

	enabled := false
	if p.Enabled != nil {
		enabled = *p.Enabled
	}
	used := true
	if p.Used != nil {
		used = *p.Used
	}

	var b Broadcaster
	b.ID = broadcasterID
	b.UserID = targetUserID
	b.DisplayName = p.DisplayName
	b.Category = p.Category
	b.CategoryColor = p.CategoryColor
	b.Bio = p.Bio
	b.Certified = p.Certified
	b.Availability = p.Availability
	b.Phone = p.Phone
	b.Services = p.Services
	b.Languages = p.Languages
	b.Attributes = p.Attributes
	b.Hours = p.Hours
	b.Photos = p.Photos
	b.Enabled = enabled
	b.Used = used
	b.CreatedAt = now
	b.UpdatedAt = now

	query := `
		INSERT INTO broadcasters (
			id, user_id, display_name, category, category_color, bio, certified,
			availability, phone, services, languages, attributes, hours, photos,
			enabled, used, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		RETURNING created_at, updated_at
	`
	err = db.QueryRow(ctx, query,
		broadcasterID, targetUserID, p.DisplayName, p.Category, p.CategoryColor, p.Bio, p.Certified,
		p.Availability, p.Phone, servicesJSON, languagesJSON, attributesJSON, hoursJSON, photosJSON,
		enabled, used, now, now,
	).Scan(&b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile creare il profilo broadcaster (potrebbe già esistere)")
	}

	return &b, nil
}

// UpdateBroadcaster aggiorna il profilo di un broadcaster (richiede autenticazione da parte dell'utente o di un admin).
//
//encore:api auth method=PUT path=/broadcasters/id/:id
func UpdateBroadcaster(ctx context.Context, id uuid.UUID, p *UpdateBroadcasterParams) (*Broadcaster, error) {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	existing, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && existing.UserID != callerID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non sei il proprietario di questo profilo broadcaster"}
	}

	if p.Attributes == nil {
		p.Attributes = existing.Attributes
	}
	if p.Hours == nil {
		p.Hours = existing.Hours
	}
	if p.Photos == nil {
		p.Photos = existing.Photos
	}

	enabled := existing.Enabled
	if p.Enabled != nil {
		enabled = *p.Enabled
	}
	used := existing.Used
	if p.Used != nil {
		used = *p.Used
	}

	alignPhotos(p.Photos)
	servicesJSON, _ := json.Marshal(p.Services)
	languagesJSON, _ := json.Marshal(p.Languages)
	attributesJSON, _ := json.Marshal(p.Attributes)
	hoursJSON, _ := json.Marshal(p.Hours)
	photosJSON, _ := json.Marshal(p.Photos)

	now := time.Now()
	query := `
		UPDATE broadcasters
		SET display_name = $1, category = $2, category_color = $3, bio = $4,
		    certified = $5, availability = $6, phone = $7, services = $8,
		    languages = $9, attributes = $10, hours = $11, photos = $12,
		    enabled = $13, used = $14, updated_at = $15
		WHERE id = $16
		RETURNING updated_at
	`
	err = db.QueryRow(ctx, query,
		p.DisplayName, p.Category, p.CategoryColor, p.Bio,
		p.Certified, p.Availability, p.Phone, servicesJSON,
		languagesJSON, attributesJSON, hoursJSON, photosJSON,
		enabled, used, now, id,
	).Scan(&existing.UpdatedAt)

	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare il profilo broadcaster")
	}

	existing.DisplayName = p.DisplayName
	existing.Category = p.Category
	existing.CategoryColor = p.CategoryColor
	existing.Bio = p.Bio
	existing.Certified = p.Certified
	existing.Availability = p.Availability
	existing.Phone = p.Phone
	existing.Services = p.Services
	existing.Languages = p.Languages
	existing.Attributes = p.Attributes
	existing.Hours = p.Hours
	existing.Photos = p.Photos
	existing.Enabled = enabled
	existing.Used = used

	return existing, nil
}

// DeleteBroadcaster rimuove il profilo broadcaster (richiede autenticazione dell'utente o di un admin).
//
//encore:api auth method=DELETE path=/broadcasters/id/:id
func DeleteBroadcaster(ctx context.Context, id uuid.UUID) error {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return err
	}

	existing, err := GetBroadcaster(ctx, id)
	if err != nil {
		return err
	}

	if !isAdmin && existing.UserID != callerID {
		return &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non puoi eliminare questo profilo broadcaster"}
	}

	_, err = db.Exec(ctx, `DELETE FROM broadcasters WHERE id = $1`, id)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "errore durante l'eliminazione del broadcaster")
	}
	return nil
}

type BroadcasterAttributesResponse struct {
	Attributes []BroadcasterAttribute `json:"attributes"`
}

type UpdateBroadcasterAttributesParams struct {
	Attributes []BroadcasterAttribute `json:"attributes"`
}

// GetBroadcasterAttributes restituisce gli attributi configurati per il broadcaster.
//
//encore:api public method=GET path=/broadcasters/id/:id/attributes
func GetBroadcasterAttributes(ctx context.Context, id uuid.UUID) (*BroadcasterAttributesResponse, error) {
	b, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Attributes == nil {
		b.Attributes = []BroadcasterAttribute{}
	}
	return &BroadcasterAttributesResponse{Attributes: b.Attributes}, nil
}

// UpdateBroadcasterAttributes aggiorna la lista di attributi del broadcaster.
//
//encore:api auth method=PUT path=/broadcasters/id/:id/attributes
func UpdateBroadcasterAttributes(ctx context.Context, id uuid.UUID, p *UpdateBroadcasterAttributesParams) (*Broadcaster, error) {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return nil, err
	}

	existing, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && existing.UserID != callerID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non sei il proprietario di questo profilo broadcaster"}
	}

	attrs := []BroadcasterAttribute{}
	if p != nil && p.Attributes != nil {
		attrs = p.Attributes
	}

	attributesJSON, _ := json.Marshal(attrs)
	now := time.Now()
	_, err = db.Exec(ctx, `UPDATE broadcasters SET attributes = $1, updated_at = $2 WHERE id = $3`, attributesJSON, now, id)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare gli attributi del broadcaster")
	}

	existing.Attributes = attrs
	existing.UpdatedAt = now
	return existing, nil
}

type BroadcasterHoursResponse struct {
	Hours []BroadcasterHours `json:"hours"`
}

type UpdateBroadcasterHoursParams struct {
	Hours []BroadcasterHours `json:"hours"`
}

// GetBroadcasterHours restituisce gli orari programmati del broadcaster.
//
//encore:api public method=GET path=/broadcasters/id/:id/hours
func GetBroadcasterHours(ctx context.Context, id uuid.UUID) (*BroadcasterHoursResponse, error) {
	b, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Hours == nil {
		b.Hours = []BroadcasterHours{}
	}
	return &BroadcasterHoursResponse{Hours: b.Hours}, nil
}

// UpdateBroadcasterHours aggiorna gli orari del broadcaster.
//
//encore:api auth method=PUT path=/broadcasters/id/:id/hours
func UpdateBroadcasterHours(ctx context.Context, id uuid.UUID, p *UpdateBroadcasterHoursParams) (*Broadcaster, error) {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return nil, err
	}

	existing, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && existing.UserID != callerID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non sei il proprietario di questo profilo broadcaster"}
	}

	hours := []BroadcasterHours{}
	if p != nil && p.Hours != nil {
		hours = p.Hours
	}

	if err := ValidateHours(hours); err != nil {
		return nil, err
	}

	hoursJSON, _ := json.Marshal(hours)
	now := time.Now()
	_, err = db.Exec(ctx, `UPDATE broadcasters SET hours = $1, updated_at = $2 WHERE id = $3`, hoursJSON, now, id)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare gli orari del broadcaster")
	}

	existing.Hours = hours
	existing.UpdatedAt = now
	return existing, nil
}

type BroadcasterPhotosResponse struct {
	Photos []BroadcasterPhoto `json:"photos"`
}

type UpdateBroadcasterPhotosParams struct {
	Photos []BroadcasterPhoto `json:"photos"`
}

// GetBroadcasterPhotos restituisce la lista di foto del broadcaster.
//
//encore:api public method=GET path=/broadcasters/id/:id/photos
func GetBroadcasterPhotos(ctx context.Context, id uuid.UUID) (*BroadcasterPhotosResponse, error) {
	b, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Photos == nil {
		b.Photos = []BroadcasterPhoto{}
	}
	return &BroadcasterPhotosResponse{Photos: b.Photos}, nil
}

// UpdateBroadcasterPhotos aggiorna la lista di foto del broadcaster.
//
//encore:api auth method=PUT path=/broadcasters/id/:id/photos
func UpdateBroadcasterPhotos(ctx context.Context, id uuid.UUID, p *UpdateBroadcasterPhotosParams) (*Broadcaster, error) {
	callerID, isAdmin, err := authedUser()
	if err != nil {
		return nil, err
	}

	existing, err := GetBroadcaster(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && existing.UserID != callerID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "permesso negato: non sei il proprietario di questo profilo broadcaster"}
	}

	photos := []BroadcasterPhoto{}
	if p != nil && p.Photos != nil {
		photos = p.Photos
	}

	alignPhotos(photos)

	for i := range photos {
		if photos[i].ImageURL != "" {
			parts := strings.Split(strings.TrimRight(photos[i].ImageURL, "/"), "/")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				if parsedUUID, err := uuid.FromString(lastPart); err == nil {
					_, _ = images.SaveImage(ctx, &images.SaveImageParams{
						UUID:        parsedUUID.String(),
						Description: photos[i].Description,
					})
				}
			}
		}
	}

	photosJSON, _ := json.Marshal(photos)
	now := time.Now()
	_, err = db.Exec(ctx, `UPDATE broadcasters SET photos = $1, updated_at = $2 WHERE id = $3`, photosJSON, now, id)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare le foto del broadcaster")
	}

	existing.Photos = photos
	existing.UpdatedAt = now
	return existing, nil
}

// GetBroadcasterImageConfig restituisce le dimensioni e proporzioni standard delle immagini dei broadcaster.
//
//encore:api public method=GET path=/broadcasters/image-config
func GetBroadcasterImageConfig(ctx context.Context) (*BroadcasterImage, error) {
	return &BroadcasterImageSize, nil
}

type CategoriesResponse struct {
	Categories []config.Category `json:"categories"`
}

// GetCategories restituisce le categorie configurate in SiteConfig.
//
//encore:api public method=GET path=/broadcasters/categories
func GetCategories(ctx context.Context) (*CategoriesResponse, error) {
	return &CategoriesResponse{
		Categories: config.SiteConfig.Categories,
	}, nil
}

type BroadcasterAttributePresetsResponse struct {
	Presets []string `json:"presets"`
}

// GetBroadcasterAttributePresets restituisce i suggerimenti rapidi per gli attributi configurati in SiteConfig.
//
//encore:api public method=GET path=/broadcasters/attribute-presets
func GetBroadcasterAttributePresets(ctx context.Context) (*BroadcasterAttributePresetsResponse, error) {
	return &BroadcasterAttributePresetsResponse{
		Presets: config.GetBroadcasterAttributes(),
	}, nil
}

type AvailabilityOptionsResponse struct {
	Options []string `json:"options"`
}

// GetAvailabilityOptions restituisce le opzioni di disponibilità configurate in SiteConfig.
//
//encore:api public method=GET path=/broadcasters/availability-options
func GetAvailabilityOptions(ctx context.Context) (*AvailabilityOptionsResponse, error) {
	return &AvailabilityOptionsResponse{
		Options: config.GetAvailabilityOptions(),
	}, nil
}

type BroadcasterServicesResponse struct {
	Services []string `json:"services"`
}

// GetBroadcasterServicesPresets restituisce i suggerimenti per i servizi configurati in SiteConfig.
//
//encore:api public method=GET path=/broadcasters/services-presets
func GetBroadcasterServicesPresets(ctx context.Context) (*BroadcasterServicesResponse, error) {
	return &BroadcasterServicesResponse{
		Services: config.GetBroadcasterServices(),
	}, nil
}


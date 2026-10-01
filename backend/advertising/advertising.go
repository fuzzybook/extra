// Service advertising gestisce l'elenco delle inserzioni pubblicitarie (Advertising Directory).
package advertising

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

var db = sqldb.NewDatabase("advertising", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Advertising struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAdvertisingParams struct {
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Status Status `json:"status,omitempty"`
}

func (p *CreateAdvertisingParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "i parametri di creazione sono obbligatori"}
	}

	email, err := normalizeEmailValue(p.Email)
	if err != nil {
		return err
	}
	p.Email = email

	nome := normalizeText(p.Nome)
	if nome == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "il nome è obbligatorio"}
	}
	p.Nome = nome

	if p.Status == "" {
		p.Status = StatusActive
	} else if p.Status != StatusActive && p.Status != StatusInactive {
		return &errs.Error{Code: errs.InvalidArgument, Message: "lo status deve essere 'active' o 'inactive'"}
	}

	return nil
}

func normalizeEmailValue(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "l'email è obbligatoria"}
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || !strings.Contains(parsed.Address, "@") || !strings.Contains(parsed.Address, ".") {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "indirizzo email non valido"}
	}
	return strings.TrimSpace(strings.ToLower(parsed.Address)), nil
}

func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.Fields(value)
	return strings.Join(parts, " ")
}

type ListAdvertisingResponse struct {
	Advertising []*Advertising `json:"advertising"`
}

// List ritorna l'intera lista delle pubblicità ordinate per data di creazione.
//
//encore:api public method=GET path=/advertising
func List(ctx context.Context) (*ListAdvertisingResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, nome, email, status, created_at
		FROM advertising
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante il recupero delle pubblicità")
	}
	defer rows.Close()

	var list []*Advertising
	for rows.Next() {
		var a Advertising
		if err := rows.Scan(&a.ID, &a.Nome, &a.Email, &a.Status, &a.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del record pubblicità")
		}
		list = append(list, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante l'iterazione delle pubblicità")
	}

	if list == nil {
		list = []*Advertising{}
	}

	return &ListAdvertisingResponse{Advertising: list}, nil
}

// Create crea un nuovo cliente/pubblicità con validazione dell'email e dei campi.
//
//encore:api public method=POST path=/advertising
func Create(ctx context.Context, p *CreateAdvertisingParams) (*Advertising, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile generare l'identificativo UUID")
	}

	item := Advertising{
		ID:     id,
		Nome:   p.Nome,
		Email:  p.Email,
		Status: p.Status,
	}

	err = db.QueryRow(ctx, `
		INSERT INTO advertising (id, nome, email, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, id, p.Nome, p.Email, p.Status).Scan(&item.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile creare la pubblicità")
	}

	return &item, nil
}

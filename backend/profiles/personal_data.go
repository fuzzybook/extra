package profiles

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// PersonalData holds the personal details linked one-to-one to a user profile.
type PersonalData struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PersonalDataParams are the editable fields of a user's personal data.
type PersonalDataParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

// GetPersonalData returns the authenticated user's own personal data.
//
//encore:api auth method=GET path=/profiles/personal-data
func GetPersonalData(ctx context.Context) (*PersonalData, error) {
	userID, err := authedUserID()
	if err != nil {
		return nil, err
	}
	pd := PersonalData{UserID: userID}
	err = db.QueryRow(ctx, `
		SELECT first_name, last_name, address, city, country, created_at, updated_at
		FROM personal_data WHERE user_id = $1
	`, userID).Scan(&pd.FirstName, &pd.LastName, &pd.Address, &pd.City, &pd.Country, &pd.CreatedAt, &pd.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "personal data not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch personal data")
	}
	return &pd, nil
}

// UpsertPersonalData creates or replaces the authenticated user's own personal data.
//
//encore:api auth method=PUT path=/profiles/personal-data
func UpsertPersonalData(ctx context.Context, p *PersonalDataParams) (*PersonalData, error) {
	userID, err := authedUserID()
	if err != nil {
		return nil, err
	}
	pd := PersonalData{
		UserID:    userID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Address:   p.Address,
		City:      p.City,
		Country:   p.Country,
	}
	err = db.QueryRow(ctx, `
		INSERT INTO personal_data (user_id, first_name, last_name, address, city, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    address = EXCLUDED.address,
		    city = EXCLUDED.city,
		    country = EXCLUDED.country,
		    updated_at = NOW()
		RETURNING created_at, updated_at
	`, userID, p.FirstName, p.LastName, p.Address, p.City, p.Country).Scan(&pd.CreatedAt, &pd.UpdatedAt)
	if isForeignKeyViolation(err) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to save personal data")
	}
	return &pd, nil
}

// authedUserID returns the UUID of the authenticated caller.
func authedUserID() (uuid.UUID, error) {
	uid, ok := auth.UserID()
	if !ok {
		return uuid.Nil, &errs.Error{Code: errs.Unauthenticated, Message: "authentication required"}
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return uuid.Nil, errs.WrapCode(err, errs.Internal, "invalid user id")
	}
	return userID, nil
}

// isForeignKeyViolation reports whether err is a Postgres foreign key constraint violation.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

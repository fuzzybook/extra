// Service extrahome provides extrahome APIs.
package extrahome

import (
	"context"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("extrahome", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type StatusResponse struct {
	Status string `json:"status"`
}

// Status reports whether the extrahome service is available.
//
//encore:api public method=GET path=/extrahome/status
func Status(ctx context.Context) (*StatusResponse, error) {
	return &StatusResponse{Status: "ok"}, nil
}

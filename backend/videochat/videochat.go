// Service videochat provides video chat APIs.
package videochat

import "context"

type StatusResponse struct {
	Status string `json:"status"`
}

// Status reports whether the videochat service is available.
//
//encore:api public method=GET path=/videochat/status
func Status(ctx context.Context) (*StatusResponse, error) {
	return &StatusResponse{Status: "ok"}, nil
}

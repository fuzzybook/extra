package videochat

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type TokenParams struct {
	RoomName string `json:"room_name"`
}

type TokenResponse struct {
	OwnerID          string `json:"owner_id"`
	Role             string `json:"role"`
	ServerURL        string `json:"server_url"`
	ParticipantToken string `json:"participant_token" encore:"sensitive"`
	RoomName         string `json:"room_name"`
	Identity         string `json:"identity"`
	ExpiresIn        int64  `json:"expires_in"`
}

// CreateToken grants the authenticated user access to an existing room by name.
//
//encore:api auth method=POST path=/videochat/token
func CreateToken(ctx context.Context, p *TokenParams) (*TokenResponse, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(p.RoomName); err != nil {
		return nil, err
	}
	client, serverURL, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	room, metadata, err := findBroadcast(ctx, client, p.RoomName)
	if err != nil {
		return nil, err
	}
	_ = logBroadcastSession(ctx, metadata.OwnerID, room.Sid, room.Name, "OPEN")
	token, err := signJoinToken(secrets.LiveKitAPIKey, secrets.LiveKitAPISecret, string(uid), p.RoomName, metadata.OwnerID == string(uid))
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate LiveKit token")
	}
	role := "viewer"
	if metadata.OwnerID == string(uid) {
		role = "owner"
	}
	return &TokenResponse{OwnerID: metadata.OwnerID, Role: role,
		ServerURL: serverURL, ParticipantToken: token, RoomName: p.RoomName,
		Identity: string(uid), ExpiresIn: int64(tokenTTL.Seconds()),
	}, nil
}

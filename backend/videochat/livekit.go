package videochat

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/twitchtv/twirp"
)

var secrets struct {
	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string
}

const tokenTTL = 10 * time.Minute
const apiTimeout = 10 * time.Second

func liveKitClient() (*lksdk.RoomServiceClient, string, error) {
	serverURL, err := websocketURL(secrets.LiveKitURL)
	if err != nil || strings.TrimSpace(secrets.LiveKitAPIKey) == "" || strings.TrimSpace(secrets.LiveKitAPISecret) == "" {
		return nil, "", &errs.Error{Code: errs.FailedPrecondition, Message: "LiveKit is not configured"}
	}
	return lksdk.NewRoomServiceClient(serverURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret), serverURL, nil
}

func websocketURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("invalid LiveKit URL")
	}
	switch u.Scheme {
	case "http", "ws":
		u.Scheme = "ws"
	case "https", "wss":
		u.Scheme = "wss"
	default:
		return "", errors.New("invalid LiveKit URL scheme")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func requireAdmin() error {
	if _, ok := auth.UserID(); !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	data, _ := auth.Data().(*authsvc.AuthData)
	if data == nil || !data.Role.IsAdmin() {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin role required"}
	}
	return nil
}

func validateRoomName(name string) error {
	if strings.TrimSpace(name) == "" || name != strings.TrimSpace(name) || len(name) > 128 || strings.ContainsAny(name, "/\\") || strings.ContainsFunc(name, unicode.IsControl) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "room name must contain 1-128 bytes, without slashes, surrounding whitespace or control characters"}
	}
	return nil
}

func liveKitError(err error) error {
	if err == nil {
		return nil
	}
	code := errs.Unavailable
	var rpcErr twirp.Error
	if errors.As(err, &rpcErr) {
		switch rpcErr.Code() {
		case twirp.NotFound:
			code = errs.NotFound
		case twirp.InvalidArgument:
			code = errs.InvalidArgument
		case twirp.AlreadyExists:
			code = errs.AlreadyExists
		}
	}
	return errs.WrapCode(err, code, "LiveKit request failed")
}

// Ownership is written only by this server when creating a fresh LiveKit room.
type broadcastMetadata struct {
	Mode        string `json:"mode"`
	OwnerID     string `json:"owner_id"`
	DisplayName string `json:"display_name"`
}

func readBroadcast(room *livekit.Room) (broadcastMetadata, bool) {
	var metadata broadcastMetadata
	err := json.Unmarshal([]byte(room.Metadata), &metadata)
	return metadata, err == nil && metadata.Mode == "broadcast" && metadata.OwnerID != "" && metadata.DisplayName != ""
}

func findBroadcast(ctx context.Context, client *lksdk.RoomServiceClient, name string) (*livekit.Room, broadcastMetadata, error) {
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{name}})
	if err != nil {
		return nil, broadcastMetadata{}, liveKitError(err)
	}
	for _, room := range result.Rooms {
		if room.Name == name {
			if metadata, ok := readBroadcast(room); ok {
				return room, metadata, nil
			}
		}
	}
	return nil, broadcastMetadata{}, &errs.Error{Code: errs.NotFound, Message: "broadcast room not found"}
}

func requireRoomOwner(ctx context.Context, client *lksdk.RoomServiceClient, name string) error {
	uid, ok := auth.UserID()
	if !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	_, metadata, err := findBroadcast(ctx, client, name)
	if err != nil {
		return err
	}
	if metadata.OwnerID != string(uid) {
		return &errs.Error{Code: errs.PermissionDenied, Message: "room owner required"}
	}
	return nil
}

func signJoinToken(key, secret, identity, room string, owner bool) (string, error) {
	grant := &lkauth.VideoGrant{RoomJoin: true, Room: room}
	grant.SetCanPublish(owner)
	grant.SetCanSubscribe(true)
	grant.SetCanPublishData(true)
	grant.SetCanUpdateOwnMetadata(false)
	return lkauth.NewAccessToken(key, secret).
		SetIdentity(identity).
		SetValidFor(tokenTTL).
		SetVideoGrant(grant).
		ToJWT()
}

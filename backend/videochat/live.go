package videochat

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"encore.app/broadcasters"
	"encore.dev/beta/errs"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// LiveStatus dice se il broadcaster della pagina pubblica /br/:shortID è in
// diretta. RoomName serve a chi ha fatto login per chiedere un token con
// CreateToken; RoomSID è la chiave della chat della diretta.
type LiveStatus struct {
	Live        bool   `json:"live"`
	DisplayName string `json:"display_name"`
	RoomName    string `json:"room_name,omitempty"`
	RoomSID     string `json:"room_sid,omitempty"`
}

// liveBroadcast trova la stanza del broadcaster con quello short_id. La
// stanza è in diretta solo se il proprietario è collegato e pubblica almeno
// una traccia; altrimenti room è nil.
func liveBroadcast(ctx context.Context, client *lksdk.RoomServiceClient, shortID string) (*broadcasters.LiveOwner, *livekit.Room, error) {
	owner, err := broadcasters.GetLiveOwner(ctx, shortID)
	if err != nil {
		if errs.Code(err) == errs.NotFound {
			return nil, nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster not found"}
		}
		return nil, nil, err
	}
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, nil, liveKitError(err)
	}
	for _, room := range result.Rooms {
		metadata, ok := readBroadcast(room)
		if !ok || metadata.OwnerID != owner.UserID {
			continue
		}
		participants, err := client.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room.Name})
		if err != nil {
			return nil, nil, liveKitError(err)
		}
		for _, p := range participants.Participants {
			if p.Identity == owner.UserID && len(p.Tracks) > 0 {
				return owner, room, nil
			}
		}
	}
	return owner, nil, nil
}

// GetLive restituisce lo stato della diretta di un broadcaster per il sito
// pubblico. Non richiede login.
//
//encore:api public method=GET path=/videochat/live/:shortID
func GetLive(ctx context.Context, shortID string) (*LiveStatus, error) {
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	owner, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	status := &LiveStatus{DisplayName: owner.DisplayName}
	if room != nil {
		status.Live = true
		status.RoomName = room.Name
		status.RoomSID = room.Sid
	}
	return status, nil
}

// CreateGuestToken permette a un visitatore senza login di guardare una
// diretta: può solo ricevere audio, video e messaggi, non pubblica nulla, non
// scrive in chat ed è nascosto agli altri partecipanti.
//
//encore:api public method=POST path=/videochat/live/:shortID/token
func CreateGuestToken(ctx context.Context, shortID string) (*TokenResponse, error) {
	client, serverURL, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	owner, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "broadcaster is not live"}
	}
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create guest identity")
	}
	identity := "guest-" + hex.EncodeToString(suffix)
	token, err := signGuestToken(secrets.LiveKitAPIKey, secrets.LiveKitAPISecret, identity, room.Name)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate LiveKit token")
	}
	return &TokenResponse{OwnerID: owner.UserID, Role: "guest",
		ServerURL: serverURL, ParticipantToken: token, RoomName: room.Name,
		Identity: identity, ExpiresIn: int64(tokenTTL.Seconds()),
	}, nil
}

// ListLiveMessages restituisce gli ultimi messaggi della chat di una diretta
// in corso, in sola lettura per i visitatori senza login.
//
//encore:api public method=GET path=/videochat/live/:shortID/messages
func ListLiveMessages(ctx context.Context, shortID string, p *ListMessagesParams) (*ListMessagesResponse, error) {
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	_, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return &ListMessagesResponse{Messages: []*ChatMessage{}}, nil
	}
	limit := 50
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}
	return recentMessages(ctx, room.Sid, limit)
}

func signGuestToken(key, secret, identity, room string) (string, error) {
	grant := &lkauth.VideoGrant{RoomJoin: true, Room: room, Hidden: true}
	grant.SetCanPublish(false)
	grant.SetCanSubscribe(true)
	grant.SetCanPublishData(false)
	grant.SetCanUpdateOwnMetadata(false)
	return lkauth.NewAccessToken(key, secret).
		SetIdentity(identity).
		SetName("Guest").
		SetValidFor(tokenTTL).
		SetVideoGrant(grant).
		ToJWT()
}

package videochat

import (
	"context"
	"strings"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/livekit/protocol/livekit"
)

type Participant struct {
	SID       string `json:"sid"`
	Identity  string `json:"identity"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type ListParticipantsResponse struct {
	Participants []*Participant `json:"participants"`
}

// ListParticipants lists participants in a room. Requires the room owner.
//
//encore:api auth method=GET path=/videochat/rooms/:roomName/participants
func ListParticipants(ctx context.Context, roomName string) (*ListParticipantsResponse, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	room, err := getOwnedBroadcast(ctx, client, roomName)
	if err != nil {
		return nil, err
	}
	result, err := client.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room.Name})
	if err != nil {
		return nil, liveKitError(err)
	}

	identities := make([]string, 0, len(result.Participants))
	for _, p := range result.Participants {
		identities = append(identities, p.Identity)
	}

	avatars := make(map[string]string)
	names := make(map[string]string)
	if len(identities) > 0 {
		rows, err := profilesDB.Query(ctx, `
			SELECT user_id::text, display_name, avatar_url
			FROM user_profiles
			WHERE user_id::text = ANY($1)
		`, identities)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var uid, dName, av string
				if err := rows.Scan(&uid, &dName, &av); err == nil {
					avatars[uid] = av
					names[uid] = dName
				}
			}
		}
	}

	response := &ListParticipantsResponse{Participants: make([]*Participant, 0, len(result.Participants))}
	for _, p := range result.Participants {
		displayName := p.Name
		if d, ok := names[p.Identity]; ok && d != "" {
			displayName = d
		}
		response.Participants = append(response.Participants, &Participant{
			SID:       p.Sid,
			Identity:  p.Identity,
			Name:      displayName,
			State:     p.State.String(),
			AvatarURL: avatars[p.Identity],
		})
	}
	return response, nil
}

// RemoveParticipant disconnects a participant. Requires the room owner.
// It does not ban the user from requesting another token.
//
//encore:api auth method=DELETE path=/videochat/rooms/:roomName/participants/:identity
func RemoveParticipant(ctx context.Context, roomName, identity string) error {
	if _, ok := auth.UserID(); !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return err
	}
	if strings.TrimSpace(identity) == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "participant identity is required"}
	}
	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	room, err := getOwnedBroadcast(ctx, client, roomName)
	if err != nil {
		return err
	}
	uid, _ := auth.UserID()
	if identity == string(uid) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "owner cannot remove themselves; leave or close the room instead"}
	}
	_, err = client.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{Room: room.Name, Identity: identity})
	return liveKitError(err)
}

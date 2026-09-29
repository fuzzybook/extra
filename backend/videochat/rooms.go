package videochat

import (
	"context"
	"encoding/json"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"

	"github.com/livekit/protocol/livekit"
)

var profilesDB = sqldb.Named("profiles")

type CreateRoomParams struct {
	Name            string `json:"name"`
	MaxParticipants uint32 `json:"max_participants"`
	EmptyTimeout    uint32 `json:"empty_timeout"`
}

type Room struct {
	Active          bool   `json:"active"`
	OwnerID         string `json:"owner_id"`
	DisplayName     string `json:"display_name"`
	Mode            string `json:"mode"`
	SID             string `json:"sid"`
	Name            string `json:"name"`
	NumParticipants uint32 `json:"num_participants"`
	MaxParticipants uint32 `json:"max_participants"`
	CreationTime    int64  `json:"creation_time"`
}

type ListRoomsResponse struct {
	Rooms []*Room `json:"rooms"`
}

func roomResponse(room *livekit.Room) *Room {
	metadata, _ := readBroadcast(room)
	return &Room{Active: true, OwnerID: metadata.OwnerID, DisplayName: metadata.DisplayName, Mode: metadata.Mode, SID: room.Sid, Name: room.Name, NumParticipants: room.NumParticipants,
		MaxParticipants: room.MaxParticipants, CreationTime: room.CreationTime}
}

// CreateRoom creates a LiveKit room. Requires an administrator.
//
//encore:api auth method=POST path=/videochat/rooms
func CreateRoom(ctx context.Context, p *CreateRoomParams) (*Room, error) {
	if err := requireAdmin(); err != nil {
		return nil, err
	}
	if err := validateRoomName(p.Name); err != nil {
		return nil, err
	}
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	uid, _ := auth.UserID()
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create room identifier")
	}
	metadata, err := json.Marshal(broadcastMetadata{Mode: "broadcast", OwnerID: string(uid), DisplayName: p.Name})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to encode room metadata")
	}
	room, err := client.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name: "broadcast-" + id.String(), Metadata: string(metadata), MaxParticipants: p.MaxParticipants, EmptyTimeout: p.EmptyTimeout,
	})
	if err != nil {
		return nil, liveKitError(err)
	}
	return roomResponse(room), nil
}

// ListRooms lists live broadcasts and enabled personal setups for authenticated viewers.
// Configured rooms absent from LiveKit are returned with active=false.
//
//encore:api auth method=GET path=/videochat/rooms
func ListRooms(ctx context.Context) (*ListRoomsResponse, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, liveKitError(err)
	}
	response := &ListRoomsResponse{Rooms: make([]*Room, 0, len(result.Rooms))}
	for _, room := range result.Rooms {
		if _, ok := readBroadcast(room); ok {
			response.Rooms = append(response.Rooms, roomResponse(room))
		}
	}
	configured, err := configuredBroadcastRooms(ctx)
	if err != nil {
		return nil, err
	}
	response.Rooms = mergeBroadcastRooms(response.Rooms, configured)
	return response, nil
}

func configuredBroadcastRooms(ctx context.Context) ([]*Room, error) {
	rows, err := profilesDB.Query(ctx, `
		SELECT user_id::text, room_name
		FROM personal_setup
		WHERE can_broadcast = TRUE AND room_name <> ''
		ORDER BY room_name, user_id
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list configured broadcast rooms")
	}
	defer rows.Close()
	rooms := make([]*Room, 0)
	for rows.Next() {
		room := &Room{Mode: "broadcast"}
		if err := rows.Scan(&room.OwnerID, &room.Name); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to read configured broadcast room")
		}
		room.DisplayName = room.Name
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list configured broadcast rooms")
	}
	return rooms, nil
}

func mergeBroadcastRooms(active, configured []*Room) []*Room {
	type ownerTitle struct{ owner, title string }
	names := make(map[string]bool, len(active))
	titles := make(map[ownerTitle]bool, len(active))
	for _, room := range active {
		names[room.Name] = true
		titles[ownerTitle{room.OwnerID, room.DisplayName}] = true
	}
	for _, room := range configured {
		if !names[room.Name] && !titles[ownerTitle{room.OwnerID, room.Name}] {
			active = append(active, room)
		}
	}
	return active
}

// DeleteRoom closes a room and disconnects its participants. Requires the room owner.
//
//encore:api auth method=DELETE path=/videochat/rooms/:roomName
func DeleteRoom(ctx context.Context, roomName string) error {
	if _, ok := auth.UserID(); !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return err
	}
	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	if err := requireRoomOwner(ctx, client, roomName); err != nil {
		return err
	}
	_, err = client.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: roomName})
	return liveKitError(err)
}

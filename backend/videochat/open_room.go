package videochat

import (
	"context"
	"encoding/json"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/livekit/protocol/livekit"
)

// OpenRoom returns an existing broadcast or provisions an enabled personal setup.
// The owner comes exclusively from administrator-managed settings, never the caller.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/open
func OpenRoom(ctx context.Context, roomName string) (*Room, error) {
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

	// Serialize lookup and creation across application instances for this name.
	tx, err := profilesDB.Begin(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to open room")
	}
	defer tx.Rollback()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "videochat/open/"+roomName); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to lock room creation")
	}
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, liveKitError(err)
	}
	for _, room := range result.Rooms {
		if room.Name == roomName {
			if _, ok := readBroadcast(room); !ok {
				return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "room name is already used by an unmanaged room"}
			}
			return roomResponse(room), nil
		}
	}

	// Multiple enabled owners with the same name are ambiguous: do not pick one.
	var owner string
	var count int
	err = tx.QueryRow(ctx, `SELECT COUNT(*), COALESCE(MIN(user_id::text), '') FROM personal_setup WHERE can_broadcast = TRUE AND room_name = $1`, roomName).Scan(&count, &owner)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch configured room")
	}
	if count == 0 {
		return nil, &errs.Error{Code: errs.NotFound, Message: "no enabled broadcast setup for this room"}
	}
	if count != 1 {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "room name is configured for more than one owner"}
	}
	for _, room := range result.Rooms {
		if metadata, ok := readBroadcast(room); ok && metadata.OwnerID == owner && metadata.DisplayName == roomName {
			return roomResponse(room), nil
		}
	}
	metadata, err := json.Marshal(broadcastMetadata{Mode: "broadcast", OwnerID: owner, DisplayName: roomName})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to encode room metadata")
	}
	room, err := client.CreateRoom(ctx, &livekit.CreateRoomRequest{Name: roomName, Metadata: string(metadata)})
	if err != nil {
		return nil, liveKitError(err)
	}
	return roomResponse(room), nil
}

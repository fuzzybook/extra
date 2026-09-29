package videochat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	authsvc "encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/proto"
)

func TestOpenRoomCreatesAndReusesConfiguredRoom(t *testing.T) {
	useTestProfilesDB(t)
	ctx := context.Background()
	const owner = "00000000-0000-4000-8000-000000000001"
	if _, err := profilesDB.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email) VALUES ($1, 'Owner', 'open@example.invalid')`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := profilesDB.Exec(ctx, `INSERT INTO personal_setup (user_id, can_broadcast, room_name) VALUES ($1, TRUE, 'Configured room')`, owner); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var rooms []*livekit.Room
	creates := 0
	failList := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var response proto.Message
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/ListRooms":
			if failList {
				w.WriteHeader(503)
				return
			}
			response = &livekit.ListRoomsResponse{Rooms: rooms}
		case "/twirp/livekit.RoomService/CreateRoom":
			body, _ := io.ReadAll(r.Body)
			var request livekit.CreateRoomRequest
			if err := proto.Unmarshal(body, &request); err != nil {
				t.Error(err)
			}
			creates++
			room := &livekit.Room{Sid: "created-sid", Name: request.Name, Metadata: request.Metadata}
			rooms = append(rooms, room)
			response = room
		default:
			t.Errorf("unexpected operation: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		payload, _ := proto.Marshal(response)
		w.Header().Set("Content-Type", "application/protobuf")
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	original := secrets
	defer func() { secrets = original; et.OverrideAuthInfo("", nil) }()
	secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = server.URL, "test-key", "test-secret"
	et.OverrideAuthInfo("", nil)
	if _, err := OpenRoom(ctx, "Configured room"); errs.Code(err) != errs.Unauthenticated {
		t.Fatalf("anonymous open: %v", err)
	}
	et.OverrideAuthInfo("viewer", &authsvc.AuthData{Role: "user"})
	if _, err := OpenRoom(ctx, "unknown"); errs.Code(err) != errs.NotFound {
		t.Fatalf("unconfigured room: %v", err)
	}
	if _, err := OpenRoom(ctx, "bad/name"); errs.Code(err) != errs.InvalidArgument {
		t.Fatalf("invalid name: %v", err)
	}

	// Concurrent viewers may provision the configured room, but never become its owner.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			et.OverrideAuthInfo("viewer", &authsvc.AuthData{Role: "user"})
			room, err := OpenRoom(ctx, "Configured room")
			if err != nil {
				t.Error(err)
				return
			}
			if !room.Active || room.OwnerID != owner || room.Name != "Configured room" || room.SID != "created-sid" {
				t.Errorf("invalid room: %+v", room)
			}
		}()
	}
	wg.Wait()
	if creates != 1 {
		t.Fatalf("expected a single creation, got %d", creates)
	}
	token, err := CreateToken(ctx, &TokenParams{RoomName: "Configured room"})
	if err != nil || token.Role != "viewer" || token.OwnerID != owner {
		t.Fatalf("opener must remain viewer: %+v, %v", token, err)
	}

	mu.Lock()
	failList = true
	mu.Unlock()
	if _, err := OpenRoom(ctx, "Configured room"); err == nil {
		t.Fatal("upstream errors must not be treated as missing rooms")
	}
	if creates != 1 {
		t.Fatal("upstream failure created a room")
	}
	mu.Lock()
	failList = false
	rooms = []*livekit.Room{{Name: "Configured room"}}
	mu.Unlock()
	if _, err := OpenRoom(ctx, "Configured room"); errs.Code(err) != errs.FailedPrecondition {
		t.Fatalf("unmanaged collision: %v", err)
	}
	mu.Lock()
	rooms = nil
	mu.Unlock()
	if _, err := profilesDB.Exec(ctx, `UPDATE personal_setup SET can_broadcast = FALSE, room_name = '' WHERE user_id = $1`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenRoom(ctx, "Configured room"); errs.Code(err) != errs.NotFound {
		t.Fatalf("disabled room was created: %v", err)
	}
}

package videochat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authsvc "encore.app/auth"
	"encore.dev/et"
	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/proto"
)

func useTestProfilesDB(t *testing.T) {
	t.Helper()
	testDB, err := et.NewTestDatabase(context.Background(), "profiles")
	if err != nil {
		t.Fatal(err)
	}
	original := profilesDB
	profilesDB = testDB
	t.Cleanup(func() { profilesDB = original })
}

func TestListRoomsIncludesInactiveSetups(t *testing.T) {
	useTestProfilesDB(t)
	ctx := context.Background()
	const owner = "00000000-0000-4000-8000-000000000001"
	const offline = "00000000-0000-4000-8000-000000000002"
	const disabled = "00000000-0000-4000-8000-000000000003"
	const sameTitle = "00000000-0000-4000-8000-000000000004"
	for _, id := range []string{owner, offline, disabled, sameTitle} {
		if _, err := profilesDB.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email) VALUES ($1, 'Test', $2)`, id, id+"@example.invalid"); err != nil {
			t.Fatal(err)
		}
	}
	for _, setup := range []struct {
		id, name string
		enabled  bool
	}{
		{owner, "Live show", true}, {offline, "Offline show", true},
		{disabled, "", false}, {sameTitle, "Live show", true},
	} {
		if _, err := profilesDB.Exec(ctx, `INSERT INTO personal_setup (user_id, can_broadcast, room_name) VALUES ($1, $2, $3)`, setup.id, setup.enabled, setup.name); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twirp/livekit.RoomService/ListRooms" {
			t.Errorf("unexpected operation: %s", r.URL.Path)
		}
		payload, err := proto.Marshal(&livekit.ListRoomsResponse{Rooms: []*livekit.Room{
			{Sid: "sid-live", Name: "broadcast-generated", NumParticipants: 3,
				Metadata: `{"mode":"broadcast","owner_id":"` + owner + `","display_name":"Live show"}`},
			{Name: "unmanaged"},
		}})
		if err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/protobuf")
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	original := secrets
	defer func() { secrets = original; et.OverrideAuthInfo("", nil) }()
	secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = server.URL, "test-key", "test-secret"
	et.OverrideAuthInfo("viewer", &authsvc.AuthData{Role: "user"})
	result, err := ListRooms(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rooms) != 3 {
		t.Fatalf("expected one live and two inactive rooms: %+v", result.Rooms)
	}
	byOwner := make(map[string]*Room)
	for _, room := range result.Rooms {
		byOwner[room.OwnerID] = room
	}
	if room := byOwner[owner]; room == nil || !room.Active || room.Name != "broadcast-generated" || room.SID != "sid-live" || room.NumParticipants != 3 {
		t.Fatalf("live room must retain its data without duplication: %+v", room)
	}
	for _, id := range []string{offline, sameTitle} {
		if room := byOwner[id]; room == nil || room.Active || room.SID != "" || room.Mode != "broadcast" || room.Name != room.DisplayName {
			t.Fatalf("expected inactive configured room: %+v", room)
		}
	}
}

func TestMergeBroadcastRoomsMatchesTechnicalName(t *testing.T) {
	live := &Room{Active: true, Name: "broadcast-123", OwnerID: "owner", DisplayName: "Title"}
	result := mergeBroadcastRooms([]*Room{live}, []*Room{{Name: "broadcast-123", OwnerID: "owner"}})
	if len(result) != 1 || result[0] != live {
		t.Fatal("technical room name must not produce a duplicate")
	}
	if got := mergeBroadcastRooms([]*Room{}, nil); got == nil || len(got) != 0 {
		t.Fatal("empty list must stay an empty JSON array")
	}
}

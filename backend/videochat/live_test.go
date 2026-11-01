package videochat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authsvc "encore.app/auth"
	"encore.app/broadcasters"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
	"encore.dev/types/uuid"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/proto"
)

func TestLiveStatusGuestTokenAndMessages(t *testing.T) {
	useTestVideochatDB(t)
	useTestProfilesDB(t)

	const ownerID = "00000000-0000-4000-8000-000000000011"
	const viewerID = "00000000-0000-4000-8000-000000000012"
	et.MockEndpoint(broadcasters.GetLiveOwner, func(ctx context.Context, shortID string) (*broadcasters.LiveOwner, error) {
		if shortID != "abc123" {
			return nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster non trovato"}
		}
		return &broadcasters.LiveOwner{UserID: ownerID, DisplayName: "Alicia"}, nil
	})
	et.MockEndpoint(broadcasters.ShortIDsByUsers, func(ctx context.Context, p *broadcasters.ShortIDsByUsersParams) (*broadcasters.ShortIDsByUsersResponse, error) {
		resp := &broadcasters.ShortIDsByUsersResponse{ShortIDs: map[string]string{}}
		for _, id := range p.UserIDs {
			if id == ownerID {
				resp.ShortIDs[id] = "abc123"
			}
		}
		return resp, nil
	})

	publishing := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response proto.Message
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/ListRooms":
			response = &livekit.ListRoomsResponse{Rooms: []*livekit.Room{
				{Sid: "RM_other", Name: "other", Metadata: `{"mode":"broadcast","owner_id":"someone-else","display_name":"Other"}`},
				{Sid: "RM_live", Name: "alicia-room", Metadata: `{"mode":"broadcast","owner_id":"` + ownerID + `","display_name":"Alicia"}`},
			}}
		case "/twirp/livekit.RoomService/ListParticipants":
			body, _ := io.ReadAll(r.Body)
			var request livekit.ListParticipantsRequest
			_ = proto.Unmarshal(body, &request)
			owner := &livekit.ParticipantInfo{Identity: ownerID}
			if publishing {
				owner.Tracks = []*livekit.TrackInfo{
					{Sid: "TR_cam", Type: livekit.TrackType_VIDEO},
					{Sid: "TR_mic", Type: livekit.TrackType_AUDIO, Muted: true},
				}
			}
			result := &livekit.ListParticipantsResponse{}
			if request.Room == "alicia-room" {
				result.Participants = []*livekit.ParticipantInfo{owner}
			}
			response = result
		case "/twirp/livekit.RoomService/SendData":
			response = &livekit.SendDataResponse{}
		default:
			t.Errorf("unexpected LiveKit operation: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload, _ := proto.Marshal(response)
		w.Header().Set("Content-Type", "application/protobuf")
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	original := secrets
	defer func() { secrets = original }()
	secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = server.URL, "test-key", "test-secret"

	ctx := context.Background()
	if _, err := GetLive(ctx, "zzz999"); errs.Code(err) != errs.NotFound {
		t.Fatalf("unknown broadcaster: %v", err)
	}

	// Stanza aperta ma il proprietario non pubblica: non è in diretta.
	status, err := GetLive(ctx, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if status.Live || status.RoomName != "" || status.DisplayName != "Alicia" {
		t.Fatalf("unexpected offline status: %+v", status)
	}
	if _, err := CreateGuestToken(ctx, "abc123"); errs.Code(err) != errs.FailedPrecondition {
		t.Fatalf("guest token while offline: %v", err)
	}
	if list, err := ListLiveMessages(ctx, "abc123", &ListMessagesParams{}); err != nil || len(list.Messages) != 0 {
		t.Fatalf("offline messages: %+v %v", list, err)
	}

	// Lo stato si mostra solo per i broadcaster con can_broadcast nel personal setup.
	publishing = true
	if list, err := ListLive(ctx); err != nil || len(list.Enabled) != 0 || len(list.Open) != 0 || len(list.Video) != 0 {
		t.Fatalf("ListLive without personal setup = %+v, %v", list, err)
	}
	publishing = false
	if _, err := profilesDB.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email) VALUES ($1, 'Alicia', 'alicia@example.invalid')`, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := profilesDB.Exec(ctx, `INSERT INTO personal_setup (user_id, can_broadcast, room_name) VALUES ($1, TRUE, 'alicia-room')`, ownerID); err != nil {
		t.Fatal(err)
	}
	// Proprietario collegato senza tracce: stanza aperta, camera e microfono spenti.
	if list, err := ListLive(ctx); err != nil || len(list.Enabled) != 1 || list.Enabled[0] != "abc123" ||
		len(list.Open) != 1 || list.Open[0] != "abc123" || len(list.Video) != 0 || len(list.Audio) != 0 {
		t.Fatalf("offline ListLive = %+v, %v", list, err)
	}

	publishing = true
	// Camera accesa, microfono in muto.
	if list, err := ListLive(ctx); err != nil || len(list.Open) != 1 || len(list.Video) != 1 || list.Video[0] != "abc123" || len(list.Audio) != 0 {
		t.Fatalf("live ListLive = %+v, %v", list, err)
	}
	status, err = GetLive(ctx, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Live || status.RoomName != "alicia-room" || status.RoomSID != "RM_live" {
		t.Fatalf("unexpected live status: %+v", status)
	}

	guest, err := CreateGuestToken(ctx, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if guest.Role != "guest" || guest.OwnerID != ownerID || guest.RoomName != "alicia-room" || !strings.HasPrefix(guest.Identity, "guest-") {
		t.Fatalf("unexpected guest token: %+v", guest)
	}
	verifier, err := lkauth.ParseAPIToken(guest.ParticipantToken)
	if err != nil {
		t.Fatal(err)
	}
	_, grants, err := verifier.Verify("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	g := grants.Video
	if grants.Identity != guest.Identity || g.Room != "alicia-room" || !g.RoomJoin || !g.Hidden ||
		g.GetCanPublish() || g.GetCanPublishData() || !g.GetCanSubscribe() || g.GetCanUpdateOwnMetadata() {
		t.Fatalf("guest must only watch: %+v", g)
	}
	if g.RoomAdmin || g.RoomCreate || g.RoomList || g.RoomRecord {
		t.Fatal("guest must not receive administrative grants")
	}

	// La chat della diretta è quella della stanza (chiave: SID), leggibile
	// senza login.
	viewerCtx := auth.WithContext(ctx, auth.UID(viewerID), &authsvc.AuthData{})
	if _, err := SendMessage(viewerCtx, "RM_live", &SendMessageParams{ClientMessageID: "c1", Text: "Ciao!"}); err != nil {
		t.Fatal(err)
	}
	list, err := ListLiveMessages(ctx, "abc123", &ListMessagesParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Messages) != 1 || list.Messages[0].Body != "Ciao!" {
		t.Fatalf("unexpected live messages: %+v", list.Messages)
	}
}

func TestOpenMyLiveUsesPersonalSetup(t *testing.T) {
	useTestVideochatDB(t)
	useTestProfilesDB(t)
	const ownerID = "00000000-0000-4000-8000-000000000021"
	const visitorID = "00000000-0000-4000-8000-000000000022"
	et.MockEndpoint(broadcasters.GetLiveOwnerByUser, func(ctx context.Context, userID uuid.UUID) (*broadcasters.LiveOwner, error) {
		if userID.String() != ownerID {
			return nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster non trovato"}
		}
		return &broadcasters.LiveOwner{UserID: ownerID, ShortID: "def456", DisplayName: "Bruna"}, nil
	})

	var rooms []*livekit.Room
	created, deleted := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response proto.Message
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/ListRooms":
			response = &livekit.ListRoomsResponse{Rooms: rooms}
		case "/twirp/livekit.RoomService/CreateRoom":
			body, _ := io.ReadAll(r.Body)
			var request livekit.CreateRoomRequest
			_ = proto.Unmarshal(body, &request)
			created++
			room := &livekit.Room{Sid: "RM_new", Name: request.Name, Metadata: request.Metadata}
			rooms = append(rooms, room)
			response = room
		case "/twirp/livekit.RoomService/DeleteRoom":
			body, _ := io.ReadAll(r.Body)
			var request livekit.DeleteRoomRequest
			_ = proto.Unmarshal(body, &request)
			deleted++
			kept := rooms[:0]
			for _, room := range rooms {
				if room.Name != request.Room {
					kept = append(kept, room)
				}
			}
			rooms = kept
			response = &livekit.DeleteRoomResponse{}
		default:
			t.Errorf("unexpected LiveKit operation: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload, _ := proto.Marshal(response)
		w.Header().Set("Content-Type", "application/protobuf")
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	original := secrets
	defer func() { secrets = original }()
	secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = server.URL, "test-key", "test-secret"

	ctx := context.Background()
	for _, p := range []struct{ id, name, email, memberType string }{
		{ownerID, "Bruna", "mylive@example.invalid", "broadcaster"},
		{visitorID, "Visitatore", "visitor@example.invalid", "visitor"},
	} {
		if _, err := profilesDB.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email, member_type) VALUES ($1, $2, $3, $4)`, p.id, p.name, p.email, p.memberType); err != nil {
			t.Fatal(err)
		}
	}

	// Utente non broadcaster: nessun comando "Apri stanza".
	visitorCtx := auth.WithContext(ctx, auth.UID(visitorID), &authsvc.AuthData{})
	if mine, err := GetMyLive(visitorCtx); err != nil || mine.IsBroadcaster || mine.CanBroadcast {
		t.Fatalf("visitor GetMyLive = %+v, %v", mine, err)
	}

	// Broadcaster senza personal setup: comando visibile, broadcasting non abilitato.
	ownerCtx := auth.WithContext(ctx, auth.UID(ownerID), &authsvc.AuthData{})
	if mine, err := GetMyLive(ownerCtx); err != nil || !mine.IsBroadcaster || mine.CanBroadcast || mine.ShortID != "def456" || mine.RoomName != "" {
		t.Fatalf("disabled GetMyLive = %+v, %v", mine, err)
	}
	if _, err := OpenMyLive(ownerCtx); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("disabled OpenMyLive: %v", err)
	}
	if created != 0 {
		t.Fatalf("CreateRoom called %d times without broadcasting enabled", created)
	}

	// Abilitato nel personal setup: la stanza ha il nome configurato e viene creata una volta sola.
	if _, err := profilesDB.Exec(ctx, `INSERT INTO personal_setup (user_id, can_broadcast, room_name) VALUES ($1, TRUE, 'Sala Bruna')`, ownerID); err != nil {
		t.Fatal(err)
	}
	if mine, err := GetMyLive(ownerCtx); err != nil || !mine.CanBroadcast || mine.RoomName != "Sala Bruna" {
		t.Fatalf("enabled GetMyLive = %+v, %v", mine, err)
	}
	for i := 0; i < 2; i++ {
		live, err := OpenMyLive(ownerCtx)
		if err != nil {
			t.Fatal(err)
		}
		if live.RoomName != "Sala Bruna" || live.RoomSID != "RM_new" {
			t.Fatalf("OpenMyLive = %+v", live)
		}
	}
	if created != 1 {
		t.Fatalf("CreateRoom called %d times, want 1", created)
	}
	metadata, ok := readBroadcast(rooms[0])
	if !ok || metadata.OwnerID != ownerID || metadata.DisplayName != "Sala Bruna" {
		t.Fatalf("room metadata = %+v", metadata)
	}

	// Il proprietario entra nella stanza appena creata come owner.
	token, err := CreateToken(ownerCtx, &TokenParams{RoomName: "Sala Bruna"})
	if err != nil || token.Role != "owner" {
		t.Fatalf("owner token = %+v, %v", token, err)
	}

	// Fine diretta: solo il proprietario chiude la stanza, e solo quella con il SID indicato.
	if err := CloseMyLive(visitorCtx, &CloseMyLiveParams{RoomSID: "RM_new"}); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("visitor CloseMyLive: %v", err)
	}
	if err := CloseMyLive(ownerCtx, &CloseMyLiveParams{RoomSID: "RM_old"}); err != nil || deleted != 0 {
		t.Fatalf("stale CloseMyLive: %v, deleted %d", err, deleted)
	}
	if err := CloseMyLive(ownerCtx, &CloseMyLiveParams{RoomSID: "RM_new"}); err != nil || deleted != 1 || len(rooms) != 0 {
		t.Fatalf("CloseMyLive: %v, deleted %d, rooms %d", err, deleted, len(rooms))
	}
	if err := CloseMyLive(ownerCtx, &CloseMyLiveParams{RoomSID: "RM_new"}); err != nil || deleted != 1 {
		t.Fatalf("CloseMyLive on a closed room: %v, deleted %d", err, deleted)
	}
	if _, err := OpenMyLive(ownerCtx); err != nil || created != 2 {
		t.Fatalf("reopen after close: %v, created %d", err, created)
	}

	// Il nome configurato occupato da una stanza di un altro utente non viene riusato.
	rooms[0].Metadata = `{"mode":"broadcast","owner_id":"someone-else","display_name":"Sala Bruna"}`
	if _, err := OpenMyLive(ownerCtx); errs.Code(err) != errs.FailedPrecondition {
		t.Fatalf("room owned by another user: %v", err)
	}

	// Broadcasting disabilitato di nuovo dall'amministrazione.
	if _, err := profilesDB.Exec(ctx, `UPDATE personal_setup SET can_broadcast = FALSE, room_name = '' WHERE user_id = $1`, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenMyLive(ownerCtx); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("re-disabled OpenMyLive: %v", err)
	}
}

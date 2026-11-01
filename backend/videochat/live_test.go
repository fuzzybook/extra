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
				owner.Tracks = []*livekit.TrackInfo{{Sid: "TR_cam"}}
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

	publishing = true
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

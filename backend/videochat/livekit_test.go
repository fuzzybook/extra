package videochat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authsvc "encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/proto"
)

func TestJoinTokenPermissions(t *testing.T) {
	token, err := signJoinToken("test-key", "test-secret", "user-123", "demo", true)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := lkauth.ParseAPIToken(token)
	if err != nil {
		t.Fatal(err)
	}
	claims, grants, err := verifier.Verify("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	if verifier.APIKey() != "test-key" || grants.Identity != "user-123" {
		t.Fatal("unexpected token issuer or identity")
	}
	g := grants.Video
	if g == nil || g.Room != "demo" || !g.RoomJoin || !g.GetCanPublish() || !g.GetCanSubscribe() || g.GetCanPublishData() || g.GetCanUpdateOwnMetadata() {
		t.Fatalf("unexpected participant grants: %+v", g)
	}
	if g.RoomAdmin || g.RoomCreate || g.RoomList || g.RoomRecord {
		t.Fatal("participant must not receive administrative grants")
	}
	if remaining := time.Until(claims.ExpiresAt.Time); remaining < tokenTTL-5*time.Second || remaining > tokenTTL {
		t.Fatalf("unexpected token lifetime: %v", remaining)
	}
	if _, _, err := verifier.Verify("wrong-secret"); err == nil {
		t.Fatal("token must reject an incorrect signing secret")
	}
}

func TestValidation(t *testing.T) {
	for _, name := range []string{"", "  ", " demo", "demo ", "a/b", "a\\b", "a\x00b", strings.Repeat("a", 129)} {
		if err := validateRoomName(name); errs.Code(err) != errs.InvalidArgument {
			t.Errorf("expected invalid room name %q, got %v", name, err)
		}
	}
	if err := validateRoomName("demo-123"); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"", "example.com", "ftp://example.com", "https://user:pass@example.com", "https://example.com?key=secret"} {
		if _, err := websocketURL(raw); err == nil {
			t.Errorf("accepted invalid URL %q", raw)
		}
	}
	for raw, want := range map[string]string{"https://example.com/": "wss://example.com", "http://localhost:7880": "ws://localhost:7880", "wss://example.com": "wss://example.com"} {
		if got, err := websocketURL(raw); err != nil || got != want {
			t.Errorf("websocketURL(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
}

func TestManagementRequiresAuthentication(t *testing.T) {
	defer et.OverrideAuthInfo("", nil)
	et.OverrideAuthInfo("", nil)
	ctx := context.Background()
	_, createErr := CreateRoom(ctx, &CreateRoomParams{Name: "demo"})
	_, listErr := ListRooms(ctx)
	deleteErr := DeleteRoom(ctx, "demo")
	_, participantsErr := ListParticipants(ctx, "demo")
	removeErr := RemoveParticipant(ctx, "demo", "user-456")
	for _, err := range []error{createErr, listErr, deleteErr, participantsErr, removeErr} {
		if errs.Code(err) != errs.Unauthenticated {
			t.Fatalf("expected unauthenticated, got %v", err)
		}
	}
	et.OverrideAuthInfo("viewer", &authsvc.AuthData{Role: "user"})
	if _, err := CreateRoom(ctx, &CreateRoomParams{Name: "demo"}); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("only administrators can create rooms: %v", err)
	}
}

func TestViewerTokenCannotPublish(t *testing.T) {
	token, err := signJoinToken("test-key", "test-secret", "viewer", "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := lkauth.ParseAPIToken(token)
	if err != nil {
		t.Fatal(err)
	}
	_, grants, err := verifier.Verify("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	g := grants.Video
	if g.GetCanPublish() || g.GetCanPublishData() || g.GetCanUpdateOwnMetadata() || !g.GetCanSubscribe() || g.RoomAdmin || g.RoomCreate {
		t.Fatalf("unsafe viewer permissions: %+v", g)
	}
}

func TestCreateTokenChecksRoomAndIdentity(t *testing.T) {
	original := secrets
	defer func() { secrets = original; et.OverrideAuthInfo("", nil) }()
	roomExists := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twirp/livekit.RoomService/ListRooms" {
			t.Errorf("unexpected LiveKit operation: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		var request livekit.ListRoomsRequest
		if err != nil || proto.Unmarshal(body, &request) != nil || len(request.Names) != 1 || request.Names[0] != "demo" {
			t.Error("room lookup must filter by the requested name")
		}
		verifier, err := lkauth.ParseAPIToken(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, grants, err := verifier.Verify("test-secret")
		if err != nil || grants.Video == nil || !grants.Video.RoomList {
			t.Error("missing signed room-list permission")
		}
		w.Header().Set("Content-Type", "application/protobuf")
		result := &livekit.ListRoomsResponse{}
		if roomExists {
			result.Rooms = []*livekit.Room{{Name: "demo", Metadata: `{"mode":"broadcast","owner_id":"owner-123","display_name":"Demo"}`}}
		}
		payload, err := proto.Marshal(result)
		if err != nil {
			t.Error(err)
		}
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	secrets.LiveKitURL = server.URL
	secrets.LiveKitAPIKey = "test-key"
	secrets.LiveKitAPISecret = "test-secret"
	ctx := context.Background()
	et.OverrideAuthInfo("", nil)
	if _, err := CreateToken(ctx, &TokenParams{RoomName: "demo"}); errs.Code(err) != errs.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	et.OverrideAuthInfo("user-123", &authsvc.AuthData{Role: "user"})
	response, err := CreateToken(ctx, &TokenParams{RoomName: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if response.Role != "viewer" || response.OwnerID != "owner-123" || response.Identity != "user-123" || response.ExpiresIn != 600 || !strings.HasPrefix(response.ServerURL, "ws://") {
		t.Fatalf("unexpected token response: %+v", response)
	}
	verifier, err := lkauth.ParseAPIToken(response.ParticipantToken)
	if err != nil {
		t.Fatal(err)
	}
	_, grants, err := verifier.Verify("test-secret")
	if err != nil || grants.Identity != "user-123" || grants.Video.Room != "demo" || grants.Video.GetCanPublish() {
		t.Fatal("token must be bound to the authenticated user and requested room")
	}
	for _, role := range []string{"user", "admin"} {
		if role == "admin" {
			et.OverrideAuthInfo("other-admin", &authsvc.AuthData{Role: "admin"})
		}
		if err := requireRoomOwner(ctx, mustTestClient(t), "demo"); errs.Code(err) != errs.PermissionDenied {
			t.Fatalf("non-owner must not manage rooms: %v", err)
		}
		if err := DeleteRoom(ctx, "demo"); errs.Code(err) != errs.PermissionDenied {
			t.Fatalf("non-owner delete: %v", err)
		}
		if _, err := ListParticipants(ctx, "demo"); errs.Code(err) != errs.PermissionDenied {
			t.Fatalf("non-owner participants: %v", err)
		}
		if err := RemoveParticipant(ctx, "demo", "viewer"); errs.Code(err) != errs.PermissionDenied {
			t.Fatalf("non-owner removal: %v", err)
		}
	}
	et.OverrideAuthInfo("owner-123", &authsvc.AuthData{Role: "admin"})
	if err := requireRoomOwner(ctx, mustTestClient(t), "demo"); err != nil {
		t.Fatal(err)
	}
	ownerResponse, err := CreateToken(ctx, &TokenParams{RoomName: "demo"})
	if err != nil || ownerResponse.Role != "owner" {
		t.Fatalf("owner token: %v", err)
	}
	ownerVerifier, _ := lkauth.ParseAPIToken(ownerResponse.ParticipantToken)
	_, ownerGrants, err := ownerVerifier.Verify("test-secret")
	if err != nil || !ownerGrants.Video.GetCanPublish() {
		t.Fatal("owner must be able to publish")
	}

	roomExists = false
	if _, err := CreateToken(ctx, &TokenParams{RoomName: "demo"}); errs.Code(err) != errs.NotFound {
		t.Fatalf("expected missing room, got %v", err)
	}
}

func TestLiveKitErrorMapping(t *testing.T) {
	for rpc, want := range map[twirp.ErrorCode]errs.ErrCode{twirp.NotFound: errs.NotFound, twirp.InvalidArgument: errs.InvalidArgument, twirp.Unavailable: errs.Unavailable} {
		if err := liveKitError(twirp.NewError(rpc, "upstream error")); errs.Code(err) != want {
			t.Errorf("expected %v, got %v", want, err)
		}
	}
}

func mustTestClient(t *testing.T) *lksdk.RoomServiceClient {
	t.Helper()
	client, _, err := liveKitClient()
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestLegacyRoomsHaveNoInferredOwner(t *testing.T) {
	for _, metadata := range []string{"", "{}", `{"mode":"broadcast"}`, `{"mode":"meeting","owner_id":"admin","display_name":"demo"}`} {
		if _, ok := readBroadcast(&livekit.Room{Metadata: metadata}); ok {
			t.Fatalf("accepted unmanaged room: %s", metadata)
		}
	}
}

// Exercise creation, listing, ownership and moderation against the real SDK's HTTP transport.
func TestBroadcastLifecycle(t *testing.T) {
	useTestProfilesDB(t)
	original := secrets
	defer func() { secrets = original; et.OverrideAuthInfo("", nil) }()
	var stored *livekit.Room
	removed := ""
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var response proto.Message
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/CreateRoom":
			var request livekit.CreateRoomRequest
			if err := proto.Unmarshal(data, &request); err != nil {
				t.Error(err)
			}
			stored = &livekit.Room{Sid: "sid-1", Name: request.Name, Metadata: request.Metadata, MaxParticipants: request.MaxParticipants}
			response = stored
		case "/twirp/livekit.RoomService/ListRooms":
			result := &livekit.ListRoomsResponse{}
			if stored != nil {
				result.Rooms = []*livekit.Room{stored, {Name: "legacy-room"}}
			}
			response = result
		case "/twirp/livekit.RoomService/ListParticipants":
			response = &livekit.ListParticipantsResponse{Participants: []*livekit.ParticipantInfo{{Sid: "viewer-sid", Identity: "viewer"}}}
		case "/twirp/livekit.RoomService/RemoveParticipant":
			var request livekit.RoomParticipantIdentity
			if err := proto.Unmarshal(data, &request); err != nil {
				t.Error(err)
			}
			removed = request.Identity
			response = &livekit.RemoveParticipantResponse{}
		case "/twirp/livekit.RoomService/DeleteRoom":
			stored = nil
			deleted = true
			response = &livekit.DeleteRoomResponse{}
		default:
			t.Errorf("unexpected operation: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/protobuf")
		payload, _ := proto.Marshal(response)
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = server.URL, "test-key", "test-secret"
	ctx := context.Background()
	et.OverrideAuthInfo("owner", &authsvc.AuthData{Role: "admin"})
	created, err := CreateRoom(ctx, &CreateRoomParams{Name: "My broadcast", MaxParticipants: 20})
	if err != nil {
		t.Fatal(err)
	}
	if created.OwnerID != "owner" || created.DisplayName != "My broadcast" || created.Mode != "broadcast" || !strings.HasPrefix(created.Name, "broadcast-") || created.MaxParticipants != 20 {
		t.Fatalf("invalid created room: %+v", created)
	}
	firstName := created.Name
	created, err = CreateRoom(ctx, &CreateRoomParams{Name: "My broadcast"})
	if err != nil || created.Name == firstName {
		t.Fatal("duplicate titles must have distinct room identities")
	}
	et.OverrideAuthInfo("viewer", &authsvc.AuthData{Role: "user"})
	listed, err := ListRooms(ctx)
	if err != nil || len(listed.Rooms) != 1 || listed.Rooms[0].OwnerID != "owner" {
		t.Fatalf("viewer must see broadcasts only: %+v, %v", listed, err)
	}
	if err := DeleteRoom(ctx, created.Name); errs.Code(err) != errs.PermissionDenied || deleted {
		t.Fatal("viewer deleted the room")
	}
	if err := RemoveParticipant(ctx, created.Name, "owner"); errs.Code(err) != errs.PermissionDenied || removed != "" {
		t.Fatal("viewer removed the owner")
	}
	et.OverrideAuthInfo("owner", &authsvc.AuthData{Role: "admin"})
	members, err := ListParticipants(ctx, created.Name)
	if err != nil || len(members.Participants) != 1 {
		t.Fatalf("owner list participants: %v", err)
	}
	if err := RemoveParticipant(ctx, created.Name, "owner"); errs.Code(err) != errs.InvalidArgument {
		t.Fatal("owner self-removal must be rejected")
	}
	if err := RemoveParticipant(ctx, created.Name, "viewer"); err != nil || removed != "viewer" {
		t.Fatalf("owner remove viewer: %v", err)
	}
	if err := DeleteRoom(ctx, created.Name); err != nil || !deleted {
		t.Fatalf("owner close: %v", err)
	}
	if _, err := CreateToken(ctx, &TokenParams{RoomName: created.Name}); errs.Code(err) != errs.NotFound {
		t.Fatal("closed room must not receive new tokens")
	}
}

package videochat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/proto"
)

func useTestVideochatDB(t *testing.T) {
	t.Helper()
	testDB, err := et.NewTestDatabase(context.Background(), "videochat")
	if err != nil {
		t.Fatal(err)
	}
	original := db
	db = testDB
	t.Cleanup(func() { db = original })
}

func mockLiveKitServer(t *testing.T, ownerID, roomName string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/ListRooms":
			payload, _ := proto.Marshal(&livekit.ListRoomsResponse{
				Rooms: []*livekit.Room{
					{
						Sid:      "RM_test123",
						Name:     roomName,
						Metadata: `{"mode":"broadcast","owner_id":"` + ownerID + `","display_name":"` + roomName + `"}`,
					},
				},
			})
			w.Header().Set("Content-Type", "application/protobuf")
			_, _ = w.Write(payload)

		case "/twirp/livekit.RoomService/UpdateParticipant":
			body, _ := io.ReadAll(r.Body)
			var req livekit.UpdateParticipantRequest
			_ = proto.Unmarshal(body, &req)

			respPayload, _ := proto.Marshal(&livekit.ParticipantInfo{
				Sid:      "PA_test",
				Identity: req.Identity,
				Permission: req.Permission,
			})
			w.Header().Set("Content-Type", "application/protobuf")
			_, _ = w.Write(respPayload)

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSpeakRequestLifecycle(t *testing.T) {
	useTestVideochatDB(t)
	useTestProfilesDB(t)

	const ownerID = "00000000-0000-4000-8000-000000000001"
	const viewerID = "00000000-0000-4000-8000-000000000002"
	const roomName = "test-intervention-room"

	server := mockLiveKitServer(t, ownerID, roomName)
	defer server.Close()

	oldURL, oldKey, oldSecret := secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret
	secrets.LiveKitURL = server.URL
	secrets.LiveKitAPIKey = "devkey"
	secrets.LiveKitAPISecret = "secret"
	defer func() {
		secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = oldURL, oldKey, oldSecret
	}()

	// 1. Viewer crea una richiesta di intervento
	viewerCtx := auth.WithContext(context.Background(), auth.UID(viewerID), &authsvc.AuthData{})

	req, err := CreateSpeakRequest(viewerCtx, roomName)
	if err != nil {
		t.Fatalf("CreateSpeakRequest failed: %v", err)
	}
	if req.Status != "pending" || req.RequesterUserID != viewerID {
		t.Fatalf("unexpected request state: %+v", req)
	}

	// 2. Idempotenza: il viewer richiama CreateSpeakRequest e ottiene la stessa richiesta
	req2, err := CreateSpeakRequest(viewerCtx, roomName)
	if err != nil {
		t.Fatalf("CreateSpeakRequest duplicate failed: %v", err)
	}
	if req2.ID != req.ID {
		t.Fatalf("expected identical request ID for duplicate, got %s vs %s", req2.ID, req.ID)
	}

	// 3. Controlla lo stato dello speaker slot
	stateResp, err := GetSpeakerState(viewerCtx, roomName)
	if err != nil {
		t.Fatalf("GetSpeakerState failed: %v", err)
	}
	if !stateResp.IsSlotAvailable || stateResp.MyRequest == nil {
		t.Fatalf("unexpected speaker state: %+v", stateResp)
	}

	// 4. Owner elenca le richieste
	ownerCtx := auth.WithContext(context.Background(), auth.UID(ownerID), &authsvc.AuthData{})

	listResp, err := ListSpeakRequests(ownerCtx, roomName, &ListSpeakRequestsParams{Status: "pending"})
	if err != nil {
		t.Fatalf("ListSpeakRequests failed: %v", err)
	}
	if len(listResp.Requests) != 1 || listResp.Requests[0].ID != req.ID {
		t.Fatalf("expected 1 pending request, got %+v", listResp.Requests)
	}

	// 5. Owner approva la richiesta
	approveResp, err := ApproveSpeakRequest(ownerCtx, roomName, req.ID)
	if err != nil {
		t.Fatalf("ApproveSpeakRequest failed: %v", err)
	}
	if !approveResp.CanPublish || approveResp.Status != "approved" {
		t.Fatalf("unexpected approve response: %+v", approveResp)
	}

	// 6. Verifica che lo slot sia ora occupato
	stateAfterApprove, err := GetSpeakerState(viewerCtx, roomName)
	if err != nil {
		t.Fatalf("GetSpeakerState after approve failed: %v", err)
	}
	if stateAfterApprove.IsSlotAvailable {
		t.Fatalf("expected slot to be occupied after approval")
	}

	// 7. Owner revoca lo speaker
	err = RevokeSpeaker(ownerCtx, roomName, viewerID)
	if err != nil {
		t.Fatalf("RevokeSpeaker failed: %v", err)
	}

	// 8. Verifica che lo slot sia di nuovo libero
	stateAfterRevoke, err := GetSpeakerState(viewerCtx, roomName)
	if err != nil {
		t.Fatalf("GetSpeakerState after revoke failed: %v", err)
	}
	if !stateAfterRevoke.IsSlotAvailable {
		t.Fatalf("expected slot to be free after revocation")
	}
}

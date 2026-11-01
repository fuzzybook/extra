package videochat

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/pubsub"
	"encore.dev/rlog"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

// RoomStatusEvent segnala che lo stato di una stanza LiveKit può essere
// cambiato (apertura, chiusura, ingresso o uscita di un partecipante, tracce).
type RoomStatusEvent struct {
	Event    string `json:"event"`
	RoomName string `json:"room_name"`
	RoomSID  string `json:"room_sid"`
}

// RoomStatusEvents riceve gli eventi dei webhook di LiveKit. È separato da
// RoomEvents (chat e comandi): qui passano solo i cambi di stato delle stanze.
var RoomStatusEvents = pubsub.NewTopic[*RoomStatusEvent]("videochat-live-status", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var _ = pubsub.NewSubscription(
	RoomStatusEvents, "update-live-status",
	pubsub.SubscriptionConfig[*RoomStatusEvent]{
		Handler: handleRoomStatusEvent,
	},
)

// Eventi che possono cambiare stanza aperta, presenza del proprietario o tracce.
var roomStatusWebhookEvents = map[string]bool{
	"room_started":                   true,
	"room_finished":                  true,
	"participant_joined":             true,
	"participant_left":               true,
	"participant_connection_aborted": true,
	"track_published":                true,
	"track_unpublished":              true,
}

// LiveKitWebhook riceve i webhook di LiveKit (configurati in livekit.yaml con
// la stessa API key del backend), ne verifica la firma e inoltra gli eventi
// delle stanze sul topic videochat-live-status.
//
//encore:api public raw method=POST path=/videochat/livekit/webhook
func LiveKitWebhook(w http.ResponseWriter, req *http.Request) {
	serveLiveKitWebhook(w, req)
}

func serveLiveKitWebhook(w http.ResponseWriter, req *http.Request) {
	if secrets.LiveKitAPIKey == "" || secrets.LiveKitAPISecret == "" {
		http.Error(w, "LiveKit is not configured", http.StatusServiceUnavailable)
		return
	}
	event, err := receiveWebhookEvent(req)
	if err != nil {
		rlog.Warn("rejected LiveKit webhook", "error", err)
		http.Error(w, "invalid webhook", http.StatusUnauthorized)
		return
	}
	room := event.GetRoom()
	if roomStatusWebhookEvents[event.GetEvent()] && room.GetName() != "" {
		if _, err := RoomStatusEvents.Publish(req.Context(), &RoomStatusEvent{
			Event:    event.GetEvent(),
			RoomName: room.GetName(),
			RoomSID:  room.GetSid(),
		}); err != nil {
			// LiveKit ritenta la consegna se la risposta non è 2xx.
			http.Error(w, "failed to publish event", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

const maxWebhookBody = 1 << 20

// receiveWebhookEvent verifica un webhook di LiveKit come webhook.Receive
// dell'SDK (senza le sue dipendenze): l'header Authorization è un token
// firmato con l'API secret che contiene lo SHA-256 del corpo.
func receiveWebhookEvent(req *http.Request) (*livekit.WebhookEvent, error) {
	data, err := io.ReadAll(http.MaxBytesReader(nil, req.Body, maxWebhookBody))
	if err != nil {
		return nil, err
	}
	token := req.Header.Get("Authorization")
	if token == "" {
		return nil, errors.New("missing authorization header")
	}
	verifier, err := lkauth.ParseAPIToken(token)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(verifier.APIKey()), []byte(secrets.LiveKitAPIKey)) != 1 {
		return nil, errors.New("unknown API key")
	}
	_, claims, err := verifier.Verify(secrets.LiveKitAPISecret)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if subtle.ConstantTimeCompare([]byte(claims.Sha256), []byte(base64.StdEncoding.EncodeToString(sum[:]))) != 1 {
		return nil, errors.New("invalid body checksum")
	}
	event := &livekit.WebhookEvent{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true, AllowPartial: true}).Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}

// handleRoomStatusEvent non applica l'evento in sé: rilegge da LiveKit lo stato
// attuale della stanza. Così eventi duplicati o fuori ordine portano comunque
// allo stato giusto. I cambi di muto non generano webhook: lo studio spegne
// camera e microfono togliendo la traccia, che invece li genera.
func handleRoomStatusEvent(ctx context.Context, ev *RoomStatusEvent) error {
	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	return refreshRoomStatus(ctx, client, ev.RoomName)
}

// refreshRoomStatus aggiorna la riga di live_status di una stanza dallo stato
// reale in LiveKit; se la stanza non esiste più (o non è broadcast) la cancella.
func refreshRoomStatus(ctx context.Context, client *lksdk.RoomServiceClient, roomName string) error {
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{roomName}})
	if err != nil {
		return liveKitError(err)
	}
	for _, room := range result.Rooms {
		if room.Name != roomName {
			continue
		}
		metadata, ok := readBroadcast(room)
		if !ok {
			break
		}
		return saveRoomStatus(ctx, client, room, metadata.OwnerID)
	}
	if _, err := db.Exec(ctx, `DELETE FROM live_status WHERE room_name = $1`, roomName); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to delete live status")
	}
	return nil
}

func saveRoomStatus(ctx context.Context, client *lksdk.RoomServiceClient, room *livekit.Room, ownerID string) error {
	participants, err := client.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room.Name})
	if err != nil {
		return liveKitError(err)
	}
	connected, video, audio := false, false, false
	for _, p := range participants.Participants {
		if p.Identity != ownerID {
			continue
		}
		connected = true
		for _, track := range p.Tracks {
			if track.Muted {
				continue
			}
			switch track.Type {
			case livekit.TrackType_VIDEO:
				video = true
			case livekit.TrackType_AUDIO:
				audio = true
			}
		}
	}
	_, err = db.Exec(ctx, `
		INSERT INTO live_status (room_name, room_sid, owner_id, owner_connected, video, audio, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (room_name) DO UPDATE SET
			room_sid = EXCLUDED.room_sid, owner_id = EXCLUDED.owner_id,
			owner_connected = EXCLUDED.owner_connected, video = EXCLUDED.video,
			audio = EXCLUDED.audio, updated_at = EXCLUDED.updated_at
	`, room.Name, room.Sid, ownerID, connected, video, audio)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to save live status")
	}
	return nil
}

var _ = cron.NewJob("sync-live-status", cron.JobConfig{
	Title:    "Realign live status with LiveKit",
	Every:    5 * cron.Minute,
	Endpoint: SyncLiveStatus,
})

// SyncLiveStatus riallinea tutta la tabella con LiveKit: recupera webhook persi
// (backend fermo, LiveKit riavviato) e toglie le stanze chiuse nel frattempo.
//
//encore:api private method=POST path=/videochat/internal/sync-live-status
func SyncLiveStatus(ctx context.Context) error {
	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return liveKitError(err)
	}
	started := time.Now()
	open := []string{}
	for _, room := range result.Rooms {
		metadata, ok := readBroadcast(room)
		if !ok {
			continue
		}
		if err := saveRoomStatus(ctx, client, room, metadata.OwnerID); err != nil {
			return err
		}
		open = append(open, room.Name)
	}
	// Le righe non aggiornate in questo giro sono stanze che non esistono più
	// (quelle aggiornate nel frattempo da un webhook restano).
	if _, err := db.Exec(ctx, `DELETE FROM live_status WHERE NOT (room_name = ANY($1)) AND updated_at < $2`, open, started); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to clean live status")
	}
	rlog.Debug("live status synced", "rooms", len(open))
	return nil
}

package videochat

import (
	"context"
	"testing"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
)

func TestChatMessages(t *testing.T) {
	useTestVideochatDB(t)
	useTestProfilesDB(t)

	const ownerID = "00000000-0000-4000-8000-000000000001"
	const senderID = "00000000-0000-4000-8000-000000000002"
	const roomName = "test-chat-room"

	server := mockLiveKitServer(t, ownerID, roomName)
	defer server.Close()

	oldURL, oldKey, oldSecret := secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret
	secrets.LiveKitURL = server.URL
	secrets.LiveKitAPIKey = "devkey"
	secrets.LiveKitAPISecret = "secret"
	defer func() {
		secrets.LiveKitURL, secrets.LiveKitAPIKey, secrets.LiveKitAPISecret = oldURL, oldKey, oldSecret
	}()

	senderCtx := auth.WithContext(context.Background(), auth.UID(senderID), &authsvc.AuthData{})

	// Invia primo messaggio
	msg1, err := SendMessage(senderCtx, roomName, &SendMessageParams{
		ClientMessageID: "msg-client-uuid-1",
		Text:            "Ciao a tutti nella stanza!",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if msg1.Body != "Ciao a tutti nella stanza!" || msg1.SenderUserID != senderID {
		t.Fatalf("unexpected message content: %+v", msg1)
	}

	// Idempotenza: invio con lo stesso client_message_id
	msgDuplicate, err := SendMessage(senderCtx, roomName, &SendMessageParams{
		ClientMessageID: "msg-client-uuid-1",
		Text:            "Ciao a tutti nella stanza!",
	})
	if err != nil {
		t.Fatalf("SendMessage duplicate failed: %v", err)
	}
	if msgDuplicate.ID != msg1.ID {
		t.Fatalf("expected identical message ID on duplicate send, got %s vs %s", msgDuplicate.ID, msg1.ID)
	}

	// Invia secondo messaggio
	_, err = SendMessage(senderCtx, roomName, &SendMessageParams{
		ClientMessageID: "msg-client-uuid-2",
		Text:            "Secondo messaggio di prova",
	})
	if err != nil {
		t.Fatalf("SendMessage 2 failed: %v", err)
	}

	// Recupera messaggi
	listResp, err := ListMessages(senderCtx, roomName, &ListMessagesParams{Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages failed: %v", err)
	}
	if len(listResp.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(listResp.Messages))
	}
	if listResp.Messages[0].Body != "Ciao a tutti nella stanza!" || listResp.Messages[1].Body != "Secondo messaggio di prova" {
		t.Fatalf("unexpected messages order or content: %+v", listResp.Messages)
	}
}

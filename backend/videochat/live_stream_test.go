package videochat

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"encore.app/broadcasters"
	"encore.dev/et"
)

func TestLiveStreamSendsStatusChanges(t *testing.T) {
	useTestVideochatDB(t)
	useTestProfilesDB(t)
	const ownerID = "00000000-0000-4000-8000-000000000041"
	et.MockEndpoint(broadcasters.ShortIDsByUsers, func(ctx context.Context, p *broadcasters.ShortIDsByUsersParams) (*broadcasters.ShortIDsByUsersResponse, error) {
		resp := &broadcasters.ShortIDsByUsersResponse{ShortIDs: map[string]string{}}
		for _, id := range p.UserIDs {
			if id == ownerID {
				resp.ShortIDs[id] = "str001"
			}
		}
		return resp, nil
	})
	poll, idle := liveStreamPoll, liveStreamIdleStop
	liveStreamPoll, liveStreamIdleStop = 50*time.Millisecond, 100*time.Millisecond
	t.Cleanup(func() { liveStreamPoll, liveStreamIdleStop = poll, idle })

	ctx := context.Background()
	if _, err := profilesDB.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email) VALUES ($1, 'Stream', 'stream@example.invalid')`, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := profilesDB.Exec(ctx, `INSERT INTO personal_setup (user_id, can_broadcast, room_name) VALUES ($1, TRUE, 'stream-room')`, ownerID); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(serveLiveStream))
	defer server.Close()
	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	req, _ := http.NewRequestWithContext(reqCtx, http.MethodGet, server.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type = %q", ct)
	}

	// Legge gli eventi "status" in arrivo.
	events := make(chan LiveList, 10)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		event := ""
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				event = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: ") && event == "status":
				var list LiveList
				if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &list) == nil {
					events <- list
				}
			}
		}
		close(events)
	}()
	waitFor := func(what string, ok func(LiveList) bool) {
		t.Helper()
		deadline := time.After(3 * time.Second)
		for {
			select {
			case list, open := <-events:
				if !open {
					t.Fatalf("stream closed waiting for %s", what)
				}
				if ok(list) {
					return
				}
			case <-deadline:
				t.Fatalf("timed out waiting for %s", what)
			}
		}
	}

	// Stato iniziale alla connessione: abilitato, stanza chiusa.
	waitFor("initial status", func(l LiveList) bool {
		return len(l.Enabled) == 1 && l.Enabled[0] == "str001" && len(l.Open) == 0
	})
	// Il webhook aggiorna la tabella: lo stream manda il nuovo stato.
	if _, err := db.Exec(ctx, `INSERT INTO live_status (room_name, room_sid, owner_id, owner_connected, video) VALUES ('stream-room', 'RM_s', $1, TRUE, TRUE)`, ownerID); err != nil {
		t.Fatal(err)
	}
	waitFor("open room with video", func(l LiveList) bool {
		return len(l.Open) == 1 && len(l.Video) == 1 && len(l.Audio) == 0
	})
	if _, err := db.Exec(ctx, `DELETE FROM live_status WHERE room_name = 'stream-room'`); err != nil {
		t.Fatal(err)
	}
	waitFor("closed room", func(l LiveList) bool { return len(l.Open) == 0 && len(l.Video) == 0 })

	// Chiusa la connessione, la lettura periodica si ferma.
	cancel()
	deadline := time.Now().Add(2 * time.Second)
	for {
		hub.mu.Lock()
		running, subs := hub.running, len(hub.subs)
		hub.mu.Unlock()
		if !running && subs == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("hub still running: running=%v subs=%d", running, subs)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

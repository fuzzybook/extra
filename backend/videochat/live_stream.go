package videochat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"encore.dev/rlog"
)

// Intervalli dello stream (variabili per i test).
var (
	liveStreamPoll      = 2 * time.Second
	liveStreamKeepalive = 20 * time.Second
	liveStreamIdleStop  = 10 * time.Second
)

// liveHub tiene una sola lettura dello stato per istanza, condivisa da tutte
// le connessioni SSE: con la tabella live_status aggiornata dai webhook (che
// arrivano a una sola istanza) ogni istanza la rilegge ogni liveStreamPoll e
// manda il nuovo stato solo quando cambia. Il carico sul database dipende dal
// numero di istanze, non dai visitatori collegati.
type liveHub struct {
	mu      sync.Mutex
	subs    map[chan []byte]struct{}
	last    []byte
	running bool
}

var hub = &liveHub{subs: map[chan []byte]struct{}{}}

// subscribe registra una connessione e avvia la lettura periodica se ferma.
// Restituisce l'ultimo stato noto (nil se non ancora letto).
func (h *liveHub) subscribe() (chan []byte, []byte) {
	ch := make(chan []byte, 1)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subs[ch] = struct{}{}
	if !h.running {
		h.running = true
		go h.run()
	}
	return ch, h.last
}

func (h *liveHub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

// run legge lo stato finché ci sono connessioni; senza connessioni per
// liveStreamIdleStop si ferma e dimentica l'ultimo stato.
func (h *liveHub) run() {
	idleSince := time.Time{}
	for {
		h.mu.Lock()
		if len(h.subs) == 0 {
			if idleSince.IsZero() {
				idleSince = time.Now()
			} else if time.Since(idleSince) >= liveStreamIdleStop {
				h.running = false
				h.last = nil
				h.mu.Unlock()
				return
			}
		} else {
			idleSince = time.Time{}
		}
		h.mu.Unlock()

		if idleSince.IsZero() {
			h.poll()
		}
		time.Sleep(liveStreamPoll)
	}
}

func (h *liveHub) poll() {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	list, err := liveList(ctx)
	if err != nil {
		rlog.Warn("live stream: failed to read status", "error", err)
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if string(data) == string(h.last) {
		return
	}
	h.last = data
	for ch := range h.subs {
		// Conta solo l'ultimo stato: chi è indietro salta quelli intermedi.
		select {
		case <-ch:
		default:
		}
		ch <- data
	}
}

// LiveStream manda lo stato delle dirette (stesso formato di ListLive) come
// Server-Sent Events: evento "status" alla connessione e a ogni cambiamento,
// commento di keepalive per non far chiudere la connessione ai proxy.
//
//encore:api public raw method=GET path=/videochat/live-stream
func LiveStream(w http.ResponseWriter, req *http.Request) {
	serveLiveStream(w, req)
}

func serveLiveStream(w http.ResponseWriter, req *http.Request) {
	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	// nginx: niente buffering della risposta.
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ch, current := hub.subscribe()
	defer hub.unsubscribe(ch)

	send := func(data []byte) bool {
		if _, err := fmt.Fprintf(w, "event: status\ndata: %s\n\n", data); err != nil {
			return false
		}
		return rc.Flush() == nil
	}
	// Prima connessione dell'istanza: stato letto subito, senza aspettare il giro.
	if current == nil {
		list, err := liveList(req.Context())
		if err == nil {
			current, _ = json.Marshal(list)
		}
	}
	if current != nil && !send(current) {
		return
	}
	// Consiglio di riconnessione per EventSource.
	if _, err := fmt.Fprint(w, "retry: 3000\n\n"); err != nil || rc.Flush() != nil {
		return
	}

	keepalive := time.NewTicker(liveStreamKeepalive)
	defer keepalive.Stop()
	for {
		select {
		case <-req.Context().Done():
			return
		case data := <-ch:
			if !send(data) {
				return
			}
		case <-keepalive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil || rc.Flush() != nil {
				return
			}
		}
	}
}

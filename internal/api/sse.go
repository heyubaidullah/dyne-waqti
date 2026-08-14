package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// keepAliveInterval is how often a comment ping is sent to idle SSE clients
// to keep intermediate proxies/browsers from closing the connection.
const keepAliveInterval = 25 * time.Second

// Broadcaster is a one-way (admin -> display) SSE fan-out hub. It has no
// buffering semantics beyond a small per-client channel: a slow or stalled
// /display client that falls behind simply misses intermediate events and
// picks up the next one, since /display always has GET /api/v1/display-data
// available to resync full state.
type Broadcaster struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{clients: make(map[chan string]struct{})}
}

func (b *Broadcaster) subscribe() chan string {
	ch := make(chan string, 4)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Publish sends an SSE event to all currently connected /display clients.
func (b *Broadcaster) Publish(event string) {
	msg := fmt.Sprintf("event: update\ndata: %s\n\n", event)
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Client is behind; drop this event for it rather than block
			// the publisher (which runs on the admin's request goroutine).
		}
	}
}

// ServeSSE streams updates to a connected /display client until it
// disconnects. No authentication is required — /display is read-only.
func (b *Broadcaster) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := b.subscribe()
	defer b.unsubscribe(ch)

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case msg := <-ch:
			if _, err := w.Write([]byte(msg)); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

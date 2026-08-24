package main

import (
	"bytes"
	"encoding/json"
	"log"
)

// sseEvent is a named server-sent event with a JSON-encodable data payload.
type sseEvent struct {
	name string
	data any
}

// encodeSSEEvent encodes an SSE event as bytes ready to write to the wire.
// The returned bytes end with the two-newline sequence that terminates an SSE event.
func encodeSSEEvent(event sseEvent) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("event: ")
	buf.WriteString(event.name)
	buf.WriteString("\ndata: ")
	if err := json.NewEncoder(&buf).Encode(event.data); err != nil {
		return nil, err
	}
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// hub manages SSE client subscriptions and event fan-out.
//
// Events are pre-encoded once as SSE message bytes before distribution.
// Clients that cannot keep up (full buffer) are silently dropped.
type hub struct {
	register   chan chan []byte
	unregister chan chan []byte
	broadcast  chan []byte
}

func newHub() *hub {
	return &hub{
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		broadcast:  make(chan []byte, 8),
	}
}

// run is the hub's main loop and must be called in its own goroutine.
func (h *hub) run() {
	clients := map[chan []byte]struct{}{}
	for {
		select {
		case ch := <-h.register:
			clients[ch] = struct{}{}
		case ch := <-h.unregister:
			if _, ok := clients[ch]; ok {
				delete(clients, ch)
				close(ch)
			}
		case msg := <-h.broadcast:
			for ch := range clients {
				select {
				case ch <- msg:
				default:
					log.Printf("hub: dropped slow client")
					delete(clients, ch)
					close(ch)
				}
			}
		}
	}
}

// subscribe returns a buffered channel that will receive encoded SSE messages.
// The caller must call unsubscribe when done.
func (h *hub) subscribe() chan []byte {
	ch := make(chan []byte, 16)
	h.register <- ch
	return ch
}

// unsubscribe removes and closes a previously subscribed channel.
// Safe to call even if the hub already dropped the client.
func (h *hub) unsubscribe(ch chan []byte) {
	h.unregister <- ch
}

// publish encodes event and broadcasts it to all subscribed clients.
func (h *hub) publish(event sseEvent) {
	msg, err := encodeSSEEvent(event)
	if err != nil {
		log.Printf("encode SSE event %q: %v", event.name, err)
		return
	}
	h.broadcast <- msg
}

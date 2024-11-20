package sse

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("sse connection not found")
)

// Manages all active SSE clients and connections
type SSEManager struct {
	clients map[string]*sseClient
}

// Creates a new SSEManager
func New() *SSEManager {
	return &SSEManager{
		clients: make(map[string]*sseClient),
	}
}

// creates a new connection for a given client ID
func (s *SSEManager) newConn(clientID string) *sseConn {
	client, ok := s.clients[clientID]
	if !ok {
		client = newClient(clientID)
		s.clients[clientID] = client
	}
	return client.newConn()
}

// removes a connection from the manager
func (s *SSEManager) removeConn(conn *sseConn) {
	client, ok := s.clients[conn.clientID]
	if !ok {
		return
	}
	client.removeConn(conn)
}

// broadcasts a message to all clients
func (s *SSEManager) Broadcast(data string) {
	for clientID := range s.clients {
		s.Publish(clientID, data)
	}
}

// sends a message to all connections of a specific client
func (s *SSEManager) Publish(clientID, data string) error {
	client, ok := s.clients[clientID]
	if !ok || client == nil {
		return ErrClientNotFound
	}

	// Send data to all client connections
	for _, conn := range client.conns {
		conn.c <- data
	}
	return nil
}

// atisfies the http.Handler interface for SSEManager
func (s SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := s.setupSSEConn(w, r)

	for {
		select {
		case data := <-conn.c:
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			slog.Info("client disconnected", "clientID", conn.clientID)
			s.removeConn(conn)
			return
		}
	}
}

// sets up an SSE connection for the incoming HTTP request
func (s *SSEManager) setupSSEConn(w http.ResponseWriter, r *http.Request) *sseConn {
	// Use client-provided ID or generate a new one
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		clientID = uuid.NewString()
	}

	// Set SSE-specific headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	return s.newConn(clientID)
}

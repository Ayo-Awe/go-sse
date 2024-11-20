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

// Represents an active SSE connection
type sseConn struct {
	c        chan string
	clientID string
	connID   string
}

// Represents a client with multiple SSE connections
type sseClient struct {
	id    string
	conns map[string]*sseConn
}

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

// Creates a new SSE client
func newClient(id string) *sseClient {
	return &sseClient{
		id:    id,
		conns: make(map[string]*sseConn),
	}
}

// Creates a new SSE connection for a given client
func newConn(clientID string) *sseConn {
	return &sseConn{
		c:        make(chan string, 1),
		clientID: clientID,
		connID:   uuid.NewString(),
	}
}

// ============================================================
// sseClient Methods
// ============================================================

// Creates and adds a new connection to the client
func (c *sseClient) newConn() *sseConn {
	conn := newConn(c.id)
	c.conns[conn.connID] = conn
	return conn
}

// Removes a connection from the client
func (c *sseClient) removeConn(conn *sseConn) {
	delete(c.conns, conn.connID)
}

// ============================================================
// SSEManager Methods
// ============================================================

// Adds a new connection for a given client ID
func (s *SSEManager) newConn(clientID string) *sseConn {
	client, ok := s.clients[clientID]
	if !ok {
		client = newClient(clientID)
		s.clients[clientID] = client
	}
	return client.newConn()
}

// Removes a connection from the manager
func (s *SSEManager) removeConn(conn *sseConn) {
	client, ok := s.clients[conn.clientID]
	if !ok {
		return
	}
	client.removeConn(conn)
}

// Broadcasts a message to all clients
func (s *SSEManager) Broadcast(data string) {
	for clientID := range s.clients {
		s.Publish(clientID, data)
	}
}

// Sends a message to all connections of a specific client
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

// Satisfies the http.Handler interface for SSEManager
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

// Sets up an SSE connection for the incoming HTTP request
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

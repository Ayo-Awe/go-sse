package sse

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("sse connection not found")
)

// An active SSE connection
type sseConn struct {
	c        chan string
	clientID string
	connID   string // uniquely identifies an sseConnection
}

type sseClient struct {
	id    string
	conns map[string]*sseConn
}

type SSEManager struct {
	clients map[string]*sseClient
}

// Constructor functions

func New() *SSEManager {
	clients := make(map[string]*sseClient)
	return &SSEManager{
		clients: clients,
	}
}

func newClient(id string) *sseClient {
	conns := make(map[string]*sseConn)
	return &sseClient{conns: conns, id: id}
}

func newConn(clientID string) *sseConn {
	c := make(chan string, 1)
	return &sseConn{
		c:        c,
		clientID: clientID,
		connID:   uuid.NewString(),
	}
}

// sseClient methods
func (c *sseClient) newConn() *sseConn {
	conn := newConn(c.id)
	c.conns[conn.connID] = conn
	return conn
}

func (c *sseClient) removeConn(conn *sseConn) {
	delete(c.conns, conn.connID)
}

// SSEManager methods
func (s *SSEManager) Broadcast(data string) {
	for clientID := range s.clients {
		s.Publish(clientID, data)
	}
}

func (s *SSEManager) newConn(clientID string) *sseConn {
	client, ok := s.clients[clientID]
	if !ok {
		client = newClient(clientID)
	}

	s.clients[clientID] = client
	return client.newConn()
}

func (s *SSEManager) Publish(clientID, data string) error {
	client, ok := s.clients[clientID]
	if !ok || client == nil {
		return ErrClientNotFound
	}

	// send data to all client connections
	for _, conn := range client.conns {
		conn.c <- data
	}

	return nil
}

func (s *SSEManager) removeConn(conn *sseConn) {
	client, ok := s.clients[conn.clientID]
	if !ok {
		return
	}

	client.removeConn(conn)
}

func (s SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := s.setupSSEConn(w, r)

	for {
		select {
		case data := <-conn.c:
			fmt.Fprintf(w, "data: %s\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			slog.Info("client disconnected")
			s.removeConn(conn)
			return
		}
	}
}

func (s *SSEManager) setupSSEConn(w http.ResponseWriter, r *http.Request) *sseConn {
	// clients can specify their id else a new client id is created
	clientID := cmp.Or(r.URL.Query().Get("id"), uuid.NewString())

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	return s.newConn(clientID)
}

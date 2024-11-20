package sse

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var (
	ErrConnNotFound = errors.New("sse connection not found")
)

// An active SSE connection
type sseConn struct {
	c  chan string
	id string
}

type SSEManager struct {
	conns map[string]*sseConn
}

func New() *SSEManager {
	conns := make(map[string]*sseConn)
	return &SSEManager{
		conns: conns,
	}
}

// creates a new connection with id, if there's an existing connection with the same id,
// it is overwritten with the new one
func newConn(id string) *sseConn {
	c := make(chan string, 1)
	return &sseConn{
		c:  c,
		id: id,
	}
}

func (s *SSEManager) addConn(conn *sseConn) {
	s.conns[conn.id] = conn
}

func (s *SSEManager) removeConn(conn *sseConn) {
	delete(s.conns, conn.id)
}

// publishes an event to all connections
func (s *SSEManager) Broadcast(data string) {
	for _, conn := range s.conns {
		conn.c <- data
	}
}

// publishes an event to a specific connection
func (s *SSEManager) Publish(id, data string) error {
	conn, ok := s.conns[id]
	if !ok || conn == nil {
		return ErrConnNotFound
	}

	conn.c <- data
	return nil
}

// satisfies http.Handler interface
func (s SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Usage: /sse?id=<client_identifier>\n")
		return
	}

	conn := s.setupSSEConn(w, id)

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

func (s *SSEManager) setupSSEConn(w http.ResponseWriter, id string) *sseConn {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	conn := newConn(id)
	s.addConn(conn)

	return conn
}

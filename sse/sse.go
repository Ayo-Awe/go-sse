package sse

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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

func NewConn() *sseConn {
	c := make(chan string, 1)
	return &sseConn{
		c:  c,
		id: uuid.New().String(),
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

// Satisfies http.Handler interface
func (s SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	conn := NewConn()
	s.addConn(conn)

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

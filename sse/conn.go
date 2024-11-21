package sse

import "github.com/google/uuid"

// Represents an active SSE connection
type sseConn struct {
	c        chan string
	clientID string
	connID   string
}

// creates a new SSE connection for a given client
func newConn(clientID string) *sseConn {
	return &sseConn{
		c:        make(chan string, 1),
		clientID: clientID,
		connID:   uuid.NewString(),
	}
}

package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	messageChannel = "sse:messages"
)

var (
	ErrClientNotFound = errors.New("sse connection not found")
)

// Manages all active SSE clients and connections
type SSEManager struct {
	clients map[string]*sseClient
	logger  *slog.Logger
	redis   *redis.Client
}

// Creates a new SSEManager
func New(ctx context.Context, redis *redis.Client) *SSEManager {
	s := &SSEManager{
		clients: make(map[string]*sseClient),
		logger:  slog.Default(),
		redis:   redis,
	}

	go s.listenForEvents(ctx)

	return s
}

// wrapper around sseClient.newConn
func (s *SSEManager) newConn(clientID string) *sseConn {
	client, ok := s.clients[clientID]
	if !ok {
		client = newClient(clientID)
		s.clients[clientID] = client
	}
	return client.newConn()
}

// wrapper around sseClient.closeConn
func (s *SSEManager) closeConn(conn *sseConn) {
	client, ok := s.clients[conn.clientID]
	if !ok {
		return
	}

	client.removeConn(conn)
	conn.close()
}

// broadcasts a message to all clients
func (s *SSEManager) Broadcast(ctx context.Context, data string) {
	// for clientID := range s.clients {
	// 	s.Publish(ctx, clientID, data)
	// }
}

// internally routes a message to all clients connected on the current process
func (s *SSEManager) publish(clientID, data string) error {
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

func (s *SSEManager) listenForEvents(ctx context.Context) {
	sub := s.redis.Subscribe(ctx, messageChannel)

	for {

		select {
		case msg := <-sub.Channel():
			s.logger.Debug("new message received")
			event := Event{}

			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				s.logger.Error("failed to unmarshal event", "err", err)
			}

			if !event.IsBroadCast() {
				s.logger.Debug("message sent to client", "id", event.ClientID)
				s.publish(event.ClientID, event.Data.(string))
			}
		case <-ctx.Done():
			// clean up resources
			s.logger.Info("no longer listening for new events")
			return
		}
	}
}

// sends a message to all connections of a specific client
func (s *SSEManager) Publish(ctx context.Context, clientID, data string) error {
	event := Event{ClientID: clientID, Data: data}

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	return s.redis.Publish(ctx, messageChannel, bytes).Err()
}

// satisfies the http.Handler interface for SSEManager
func (s SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := s.setupSSEConn(w, r)
	defer s.closeConn(conn)

	s.logger.Info("client connected", "clientID", conn.clientID)

	for {
		select {
		case data := <-conn.c:
			fmt.Fprintf(w, "data: %s\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			s.logger.Info("client disconnected", "clientID", conn.clientID)
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

package sse

// represents a client with multiple SSE connections
type sseClient struct {
	id    string
	conns map[string]*sseConn
}

// creates a new SSE client
func newClient(id string) *sseClient {
	return &sseClient{
		id:    id,
		conns: make(map[string]*sseConn),
	}
}

// creates and adds a new connection to the client
func (c *sseClient) newConn() *sseConn {
	conn := newConn(c.id)
	c.conns[conn.connID] = conn
	return conn
}

// removes a connection from the client
func (c *sseClient) removeConn(conn *sseConn) {
	delete(c.conns, conn.connID)
}

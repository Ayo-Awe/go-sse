# Go Server-Sent Events (SSE)

## Overview

This application implements a simple **Server-Sent Events (SSE)** system using Go. It allows clients to establish real-time connections to the server for receiving updates. The server can broadcast messages to all connected clients or send targeted messages to specific clients based on their unique identifiers. A single client can establish multiple connections, such as having multiple browser tabs open, and all connections will receive the same message.

## Features

- **Broadcast Messages**: Send messages to all connected clients simultaneously.
- **Targeted Messages**: Deliver messages to all connections associated with a specific client identified by a unique ID.

## Prerequisites

- [Go](https://golang.org/dl/) installed on your machine.
- Basic knowledge of running Go applications.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/ayo-awe/go-sse.git
   cd go-sse
   ```

2. Install the dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Start the Server

Run the server in a terminal:

```bash
go run .
```

### Connect a Client

To establish an SSE connection to the server, open a new terminal and use `curl` to request the SSE endpoint:

```bash
curl -N "http://localhost:3020/sse?id=<unique_id>"
```

- **`id`**: (Optional) A unique identifier for the client. If provided, it enables the server to associate multiple connections with the same client (e.g., multiple browser tabs for a single user).

### Publishing a message to a specific client

The SSEManager has a `Publish` method that can be used to send a message to a specific client. This demo exposes a `/ping/<id>` endpoint that publishes a hello message to all connections associated with a specific client ID:

```bash
curl -X POST "http://localhost:3020/ping/<id>"
```

- Replace `<id>` with the target client's unique identifier.

### Broadcasting a message to all clients

The SSEManager has a `Broadcast` method that can be used to send a message to all connected clients. In this demo, the server broadcasts the current time to all connected clients every second:

## Notes

- Multiple connections for the same client ID will all receive the same messages.

## Future Work

- Add graceful shutdown
- Make the SSEManager thread-safe

# Limitations

- Connections are maintained in memory and are not persisted between restarts.
- Not suitable for multiple instances of the server.

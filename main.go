package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case currentTime := <-ticker.C:
			fmt.Fprintf(w, "data: %s\n", currentTime.String())
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			slog.Info("client disconnected")
			return
		}
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", handleSSE)

	port := 3020
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("failed to start server", "err", err, "port", port)
	}
}

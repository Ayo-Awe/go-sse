package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/awe-ayo/go-sse/sse"
)

func main() {
	sse := sse.New()

	mux := http.NewServeMux()
	mux.Handle("/sse", sse)

	mux.HandleFunc("POST /broadcast", func(w http.ResponseWriter, r *http.Request) {
		sse.Broadcast("Hello World")
		w.WriteHeader(http.StatusOK)
	})

	port := 3020
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("failed to start server", "err", err, "port", port)
	}
}

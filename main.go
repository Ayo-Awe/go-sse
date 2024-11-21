package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/awe-ayo/go-sse/sse"
)

func main() {
	sse := sse.New()

	mux := http.NewServeMux()
	mux.Handle("/sse", sse)

	mux.HandleFunc("POST /ping/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		msg := fmt.Sprintf("Hello %s", id)

		if err := sse.Publish(id, msg); err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "failed to publish event")
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for currentTime := range ticker.C {
			sse.Broadcast(currentTime.Format(time.DateTime))
		}
	}()

	port := 3020
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("failed to start server", "err", err, "port", port)
	}
}

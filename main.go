package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/awe-ayo/go-sse/sse"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := flag.Int("port", 3020, "any available port")
	flag.Parse()

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	sse := sse.New(ctx, rdb)

	mux := http.NewServeMux()
	mux.Handle("/sse", sse)

	mux.HandleFunc("POST /ping/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		msg := fmt.Sprintf("Hello %s", id)

		if err := sse.Publish(r.Context(), id, msg); err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "failed to publish event")
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	slog.Info("starting server", "port", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), mux); err != nil {
		slog.Error("failed to start server", "err", err, "port", *port)
	}
}

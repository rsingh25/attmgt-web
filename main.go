package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

var log *slog.Logger

func init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", HelloWorldHandler)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info("Staring HTTP server", "port", 8080)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Error("Server failed to start", "err", err)
	}

}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("encode error", "error", err.Error(), "method", r.Method, "url", r.URL, "stack", debug.Stack())
		return
	}
}

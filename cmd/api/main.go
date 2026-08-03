package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codersgyan/olx-api/internal/config"
)

func main() {

	cfg := config.MustLoad()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	fmt.Printf("Server is listening on port %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

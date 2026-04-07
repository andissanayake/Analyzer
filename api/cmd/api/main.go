package main

import (
	"log"
	"net/http"

	"analyzer/api/internal/features/analyze"
	"analyzer/api/internal/features/health"
	"analyzer/api/internal/platform/config"
	"analyzer/api/internal/platform/httpx"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	health.Register(mux)
	analyze.Register(mux)

	addr := ":" + cfg.Port
	log.Printf("api listening on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, httpx.WithCORS(mux, cfg.CORSOrigin)); err != nil {
		log.Fatal(err)
	}
}

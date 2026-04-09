package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"analyzer/api/internal/features/analyze"
	"analyzer/api/internal/features/health"
	"analyzer/api/internal/platform/config"
	"analyzer/api/internal/platform/httpx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	mux := http.NewServeMux()
	health.Register(mux)

	httpClient := analyze.NewLiveHTTPClient(&http.Client{Timeout: 15 * time.Second})
	analyzeService := analyze.NewService(httpClient)
	analyze.Register(mux, analyzeService)

	addr := ":" + cfg.Port
	slog.Info("api starting", "addr", addr, "cors_origin", cfg.CORSOrigin)

	if err := http.ListenAndServe(addr, httpx.WithCORS(mux, cfg.CORSOrigin)); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

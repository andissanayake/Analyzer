package main

import (
	"log/slog"
	"net/http"
	"net/http/pprof"
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
	if cfg.PprofOn {
		debugMux := http.NewServeMux()
		debugMux.HandleFunc("/debug/pprof/", pprof.Index)
		debugMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		debugMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		debugMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		debugMux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
		debugMux.Handle("/debug/pprof/block", pprof.Handler("block"))
		debugMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		debugMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		debugMux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
		debugMux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

		go func() {
			slog.Info("pprof debug server starting", "addr", cfg.PprofAddr)
			if err := http.ListenAndServe(cfg.PprofAddr, debugMux); err != nil {
				slog.Error("pprof debug server stopped", "error", err)
			}
		}()
	}

	slog.Info("api starting", "addr", addr, "cors_origin", cfg.CORSOrigin, "pprof_enabled", cfg.PprofOn, "pprof_addr", cfg.PprofAddr)

	if err := http.ListenAndServe(addr, httpx.WithCORS(mux, cfg.CORSOrigin)); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

package analyze

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"analyzer/api/internal/platform/httpx"
	"analyzer/api/internal/platform/metrics"
)

func Register(mux *http.ServeMux, svc Service) {
	if svc == nil {
		panic("analyze: Service is required")
	}
	h := &handler{service: svc}
	mux.HandleFunc("/analyze", h.handle)
}

type handler struct {
	service Service
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	analyzedURL := ""
	appStatus := http.StatusOK
	errorClass := "none"
	defer func() {
		metrics.ObserveAnalyzeRequest(appStatus, time.Since(start).Seconds())
		slog.Info("analyze request",
			"url", analyzedURL,
			"duration_ms", time.Since(start).Milliseconds(),
			"status_code", appStatus,
			"error_class", errorClass,
		)
	}()

	if r.Method != http.MethodPost {
		appStatus = http.StatusMethodNotAllowed
		errorClass = "method"
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appStatus = http.StatusBadRequest
		errorClass = "validation"
		httpx.WriteJSON(w, http.StatusOK, analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request body",
		})
		return
	}
	analyzedURL = req.URL

	result := h.service.Analyze(r.Context(), req.URL)
	appStatus = result.StatusCode
	errorClass = classifyAnalyzeError(result.StatusCode)
	httpx.WriteJSON(w, http.StatusOK, result)
}

func classifyAnalyzeError(status int) string {
	switch {
	case status < http.StatusBadRequest:
		return "none"
	case status < http.StatusInternalServerError:
		return "validation"
	case status == http.StatusBadGateway || status == http.StatusGatewayTimeout:
		return "upstream"
	default:
		return "internal"
	}
}

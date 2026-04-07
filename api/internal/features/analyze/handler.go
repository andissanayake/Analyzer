package analyze

import (
	"encoding/json"
	"net/http"

	"analyzer/api/internal/platform/httpx"
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusOK, analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request body",
		})
		return
	}

	result := h.service.Analyze(r.Context(), req.URL)
	httpx.WriteJSON(w, http.StatusOK, result)
}

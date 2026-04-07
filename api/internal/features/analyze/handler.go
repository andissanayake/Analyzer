package analyze

import (
	"encoding/json"
	"errors"
	"net/http"

	"analyzer/api/internal/platform/httpx"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/analyze", handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request body",
		})
		return
	}

	result, err := analyzeURL(req.URL)
	if err != nil {
		status := http.StatusBadGateway
		message := "failed to reach the requested URL"
		if errors.Is(err, errURLNotReachable) {
			status = http.StatusBadRequest
			message = "the provided URL is not reachable"
		}

		httpx.WriteJSON(w, status, errorResponse{
			StatusCode: status,
			Message:    message,
		})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

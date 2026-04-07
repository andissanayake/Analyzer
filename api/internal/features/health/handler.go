package health

import (
	"net/http"

	"analyzer/api/internal/platform/httpx"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response{Status: getStatus()})
}

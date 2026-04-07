package analyze

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAnalyze_RemoteHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		remote   int
		wantApp  int
		wantBody bool
	}{
		{"remote 200 OK", http.StatusOK, http.StatusOK, true},
		{"remote 404", http.StatusNotFound, http.StatusNotFound, false},
		{"remote 500", http.StatusInternalServerError, http.StatusInternalServerError, false},
		{"remote 302 redirect", http.StatusFound, http.StatusFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remote := tt.remote
			svc := NewService(&mockHTTPClient{
				getFunc: func(ctx context.Context, url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: remote,
						Body:       io.NopCloser(strings.NewReader("<html/>")),
					}, nil
				},
			})

			result := svc.Analyze(context.Background(), "https://example.com")
			if result.StatusCode != tt.wantApp {
				t.Fatalf("status: want %d, got %d", tt.wantApp, result.StatusCode)
			}
			if tt.wantBody {
				if result.Body == nil {
					t.Fatal("expected analysis body")
				}
				return
			}
			if result.Body != nil {
				t.Fatal("expected no body")
			}
		})
	}
}

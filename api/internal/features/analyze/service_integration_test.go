package analyze

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAnalyze_Integration_LiveHTTPClient_Success(t *testing.T) {
	const realURL = "https://example.com"

	httpClient := &http.Client{Timeout: 8 * time.Second}
	svc := NewService(NewLiveHTTPClient(httpClient))

	result := svc.Analyze(context.Background(), realURL)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if result.Body == nil {
		t.Fatal("expected analysis body")
	}
	if result.Body.HTMLVersion == "" {
		t.Fatal("expected HTMLVersion to be parsed")
	}
	if result.Body.PageTitle == "" {
		t.Fatal("expected non-empty page title")
	}
	if len(result.Body.Headings) != 6 {
		t.Fatalf("expected 6 heading counters, got %d", len(result.Body.Headings))
	}
}

func TestAnalyze_Integration_LiveHTTPClient_LoginPages(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "github login", url: "https://github.com/login"},
		{name: "linkedin login", url: "https://www.linkedin.com/login"},
	}

	httpClient := &http.Client{Timeout: 12 * time.Second}
	svc := NewService(NewLiveHTTPClient(httpClient))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.Analyze(context.Background(), tt.url)
			if result.StatusCode != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, result.StatusCode)
			}
			if result.Body == nil {
				t.Fatal("expected analysis body")
			}
			if result.Body.LoginScore < 70 {
				t.Fatalf("expected login score >= 70, got %d (%s)", result.Body.LoginScore, result.Body.LoginReason)
			}
			if !result.Body.HasLoginForm {
				t.Fatalf("expected HasLoginForm true, got false (score=%d)", result.Body.LoginScore)
			}
			if !strings.Contains(strings.ToLower(result.Body.LoginReason), "password") {
				t.Fatalf("expected login reason to mention password signal, got %q", result.Body.LoginReason)
			}
		})
	}
}

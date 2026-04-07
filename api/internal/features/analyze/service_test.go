package analyze

import (
	"errors"
	"net/http"
	"testing"
)

func TestAnalyzeURL_EmptyURL_ReturnsReachableError(t *testing.T) {
	_, err := analyzeURL("")
	if !errors.Is(err, errURLNotReachable) {
		t.Fatalf("expected errURLNotReachable, got %v", err)
	}
}

func TestAnalyzeURL_ReachableURL_ReturnsAnalysis(t *testing.T) {
	result, err := analyzeURL("https://www.google.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if result.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if result.Body.HTMLVersion == "" {
		t.Fatal("expected Body.HTMLVersion to be set")
	}
}
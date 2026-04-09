package analyze

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	getFunc func(ctx context.Context, url string) (*http.Response, error)
	headFunc func(ctx context.Context, url string) (*http.Response, error)
}

func (m *mockHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	return m.getFunc(ctx, url)
}

func (m *mockHTTPClient) Head(ctx context.Context, url string) (*http.Response, error) {
	if m.headFunc == nil {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}
	return m.headFunc(ctx, url)
}

func TestAnalyze_EmptyURL_ReturnsBadRequest(t *testing.T) {
	svc := NewService(&mockHTTPClient{
		getFunc: func(ctx context.Context, url string) (*http.Response, error) {
			t.Fatal("expected no HTTP request for empty url")
			return nil, nil
		},
	})

	result := svc.Analyze(context.Background(), "")
	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, result.StatusCode)
	}
	if result.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if result.Body != nil {
		t.Fatal("expected no body for validation error")
	}
}

func TestAnalyze_ReachableURL_ReturnsAnalysis(t *testing.T) {
	svc := NewService(&mockHTTPClient{
		getFunc: func(ctx context.Context, url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
			}, nil
		},
	})

	result := svc.Analyze(context.Background(), "https://example.com")
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if result.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if result.Body == nil || result.Body.HTMLVersion == "" {
		t.Fatal("expected Body with HTMLVersion")
	}
}

func TestAnalyze_HTTPClientError_ReturnsBadGateway(t *testing.T) {
	svc := NewService(&mockHTTPClient{
		getFunc: func(ctx context.Context, url string) (*http.Response, error) {
			return nil, fmt.Errorf("connection refused")
		},
	})

	result := svc.Analyze(context.Background(), "https://example.com")
	if result.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, result.StatusCode)
	}
	if result.Message == "" || !strings.Contains(result.Message, "connection refused") {
		t.Fatalf("expected message about connection error, got %q", result.Message)
	}
	if result.Body != nil {
		t.Fatal("expected no body")
	}
}

package analyze

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Service interface {
	Analyze(ctx context.Context, url string) analysisResult
}

type service struct {
	client HTTPClient
}

func NewService(client HTTPClient) Service {
	if client == nil {
		panic("analyze: HTTPClient is required")
	}
	return &service{client: client}
}
const (
	MaxHTMLBytes           = 20_971_520 // 20MB
	LinkCheckWorkerCount   = 8
	LinkCheckTimeoutPerURL = 3 * time.Second
)

func (s *service) Analyze(ctx context.Context, rawURL string) analysisResult {
	slog.Debug("analyze started", "url", rawURL)

	if rawURL == "" {
		slog.Warn("analyze validation failed", "url", rawURL, "reason", "missing_url")
		return analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "A URL is required.",
		}
	}

	pageURL, err := url.Parse(rawURL)
	if err != nil || pageURL.Scheme == "" || pageURL.Host == "" {
		slog.Warn("analyze validation failed", "url", rawURL, "reason", "invalid_url")
		return analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "The URL is not valid.",
		}
	}

	resp, err := s.client.Get(ctx, rawURL)
	if err != nil {
		slog.Warn("analyze upstream fetch failed", "url", rawURL, "error", err)
		return analysisResult{
			StatusCode: http.StatusBadGateway,
			Message:    fmt.Sprintf("Could not reach the URL: %v", err),
		}
	}
	if resp.Body != nil {
		defer func() {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		slog.Warn("analyze upstream returned non-success", "url", rawURL, "upstream_status", resp.StatusCode)
		return analysisResult{
			StatusCode: resp.StatusCode,
			Message:    nonSuccessHTTPMessage(resp.StatusCode),
		}
	}

	limited := io.LimitReader(io.Reader(resp.Body), MaxHTMLBytes)
	payload, linkTargets, err := parseHTMLWithLinks(pageURL, limited)
	if err != nil {
		slog.Warn("analyze parse failed", "url", rawURL, "error", err)
		return analysisResult{
			StatusCode: http.StatusBadGateway,
			Message:    fmt.Sprintf("Could not analyze page content: %v", err),
		}
	}
	inaccessibleCount := s.countInaccessibleLinks(ctx, linkTargets)
	payload.InaccessibleLinks += inaccessibleCount
	slog.Info("analyze link check summary",
		"url", rawURL,
		"checked", len(uniqueHTTPLinks(linkTargets)),
		"inaccessible", inaccessibleCount,
	)

	slog.Debug("analyze completed", "url", rawURL)
	return analysisResult{
		StatusCode: http.StatusOK,
		Message:    "Analysis complete.",
		Body:       &payload,
	}
}

func nonSuccessHTTPMessage(status int) string {
	text := http.StatusText(status)
	return fmt.Sprintf("The page returned HTTP %d %s.", status, text)
}

package analyze

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (s *service) Analyze(ctx context.Context, rawURL string) analysisResult {
	if rawURL == "" {
		return analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "A URL is required.",
		}
	}

	pageURL, err := url.Parse(rawURL)
	if err != nil || pageURL.Scheme == "" || pageURL.Host == "" {
		return analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "The URL is not valid.",
		}
	}

	resp, err := s.client.Get(ctx, rawURL)
	if err != nil {
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
		return analysisResult{
			StatusCode: resp.StatusCode,
			Message:    nonSuccessHTTPMessage(resp.StatusCode),
		}
	}

	limited := io.LimitReader(resp.Body, MaxHTMLBytes)
	payload, err := ParseHTML(pageURL, limited)
	if err != nil {
		return analysisResult{
			StatusCode: http.StatusBadGateway,
			Message:    fmt.Sprintf("Could not analyze page content: %v", err),
		}
	}

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

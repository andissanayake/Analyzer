package analyze

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

func (s *service) Analyze(ctx context.Context, url string) analysisResult {
	if url == "" {
		return analysisResult{
			StatusCode: http.StatusBadRequest,
			Message:    "A URL is required.",
		}
	}

	resp, err := s.client.Get(ctx, url)
	if err != nil {
		return analysisResult{
			StatusCode: http.StatusBadGateway,
			Message:    fmt.Sprintf("Could not reach the URL: %v", err),
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
		}
		return analysisResult{
			StatusCode: resp.StatusCode,
			Message:    nonSuccessHTTPMessage(resp.StatusCode),
		}
	}

	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	return analysisResult{
		StatusCode: http.StatusOK,
		Message:    "Analysis complete.",
		Body: &analysisPayload{
			HTMLVersion: "HTML5",
			PageTitle:   "Placeholder title",
			Headings: []headingCount{
				{Level: "h1", Count: 1},
				{Level: "h2", Count: 2},
				{Level: "h3", Count: 0},
				{Level: "h4", Count: 0},
				{Level: "h5", Count: 0},
				{Level: "h6", Count: 0},
			},
			InternalLinks:     0,
			ExternalLinks:     0,
			InaccessibleLinks: 0,
			HasLoginForm:      false,
		},
	}
}

func nonSuccessHTTPMessage(status int) string {
	text := http.StatusText(status)
	return fmt.Sprintf("The page returned HTTP %d %s.", status, text)
}

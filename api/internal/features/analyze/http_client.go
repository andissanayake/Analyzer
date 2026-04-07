package analyze

import (
	"context"
	"net/http"
)

type HTTPClient interface {
	Get(ctx context.Context, url string) (*http.Response, error)
}

type liveHTTPClient struct {
	client *http.Client
}

func NewLiveHTTPClient(client *http.Client) HTTPClient {
	if client == nil {
		panic("analyze: http.Client is required")
	}

	return &liveHTTPClient{client: client}
}

func (c *liveHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
}

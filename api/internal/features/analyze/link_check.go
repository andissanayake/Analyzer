package analyze

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

func parseHTMLWithLinks(pageURL *url.URL, body io.Reader) (analysisPayload, []string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return analysisPayload{}, nil, err
	}

	return analyzeDocumentWithLinks(doc, pageURL)
}

func uniqueHTTPLinks(linkTargets []string) []string {
	unique := make([]string, 0, len(linkTargets))
	seen := make(map[string]struct{}, len(linkTargets))

	for _, target := range linkTargets {
		parsed, err := url.Parse(target)
		if err != nil {
			continue
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			continue
		}

		normalized := parsed.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		unique = append(unique, normalized)
	}

	return unique
}

func (s *service) countInaccessibleLinks(ctx context.Context, linkTargets []string) int {
	links := uniqueHTTPLinks(linkTargets)
	if len(links) == 0 {
		return 0
	}

	workerCount := LinkCheckWorkerCount
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(links) {
		workerCount = len(links)
	}

	jobs := make(chan string)
	results := make(chan int, len(links))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for target := range jobs {
			results <- s.checkLink(ctx, target)
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	for _, target := range links {
		jobs <- target
	}
	close(jobs)

	wg.Wait()
	close(results)

	total := 0
	for v := range results {
		total += v
	}

	return total
}

func (s *service) checkLink(ctx context.Context, target string) int {
	reqCtx, cancel := context.WithTimeout(ctx, LinkCheckTimeoutPerURL)
	resp, err := s.client.Get(reqCtx, target)
	cancel()
	if err != nil {
		return 1
	}

	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return 1
	}
	return 0
}

package analyze

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseHTML_DetectsHTML5Doctype(t *testing.T) {
	pageURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader("<!DOCTYPE html><html><head><title>Hello</title></head><body></body></html>")
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.HTMLVersion != "HTML5" {
		t.Fatalf("expected HTMLVersion HTML5, got %q", got.HTMLVersion)
	}
}

func TestParseHTML_DetectsHTML401Doctype(t *testing.T) {
	pageURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"><html><body></body></html>`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.HTMLVersion != "HTML 4.01" {
		t.Fatalf("expected HTMLVersion HTML 4.01, got %q", got.HTMLVersion)
	}
}

func TestParseHTML_WithoutDoctype_DefaultsToHTML5(t *testing.T) {
	pageURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader("<html><head><title>No doctype</title></head><body></body></html>")
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.HTMLVersion != "HTML5" {
		t.Fatalf("expected HTMLVersion HTML5 fallback, got %q", got.HTMLVersion)
	}
}

func TestParseHTML_ClassifiesAnchorLinks(t *testing.T) {
	pageURL, err := url.Parse("https://example.com/path")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`
		<html><body>
			<a href="/internal">internal</a>
			<a href="https://example.com/also-internal">internal absolute</a>
			<a href="https://other.example/out">external</a>
			<a href="">empty</a>
			<a href="mailto:test@example.com">mail</a>
		</body></html>
	`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.InternalLinks != 2 {
		t.Fatalf("expected 2 internal links, got %d", got.InternalLinks)
	}
	if got.ExternalLinks != 1 {
		t.Fatalf("expected 1 external link, got %d", got.ExternalLinks)
	}
	if got.InaccessibleLinks != 2 {
		t.Fatalf("expected 2 inaccessible links, got %d", got.InaccessibleLinks)
	}
}

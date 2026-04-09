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

func TestParseHTML_LoginPage_HighScoreAndHasLoginForm(t *testing.T) {
	pageURL, err := url.Parse("https://example.com/login")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Sign in to your account</title>
			<meta name="description" content="Use your username and password">
		</head>
		<body>
			<form action="/auth/login" method="post">
				<input type="text" name="username" autocomplete="username">
				<input type="password" name="password" autocomplete="current-password">
				<input type="submit" value="Sign in">
			</form>
			<a href="/forgot-password">Forgot password?</a>
		</body>
		</html>
	`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.LoginScore < 70 {
		t.Fatalf("expected login score >= 70, got %d (reason: %s)", got.LoginScore, got.LoginReason)
	}
	if !got.HasLoginForm {
		t.Fatalf("expected HasLoginForm true when score >= 70, got false (score=%d)", got.LoginScore)
	}
	if got.PageTitle != "Sign in to your account" {
		t.Fatalf("expected page title from parser, got %q", got.PageTitle)
	}
	if got.LoginReason == "" {
		t.Fatal("expected non-empty LoginReason")
	}
}

func TestParseHTML_LoginPage_MediumScore_NoHasLoginForm(t *testing.T) {
	pageURL, err := url.Parse("https://example.com/session")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`
		<html>
		<head><title>Account access</title></head>
		<body>
			<form action="/session">
				<input type="password" name="passcode">
				<input type="submit" value="Continue">
			</form>
		</body>
		</html>
	`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.LoginScore < 40 || got.LoginScore > 69 {
		t.Fatalf("expected medium login score in [40,69], got %d (reason: %s)", got.LoginScore, got.LoginReason)
	}
	if got.HasLoginForm {
		t.Fatalf("expected HasLoginForm false when score < 70, got true (score=%d)", got.LoginScore)
	}
	if !strings.Contains(strings.ToLower(got.LoginReason), "password") {
		t.Fatalf("expected LoginReason to mention password signal, got %q", got.LoginReason)
	}
}

func TestParseHTML_NonLoginContactForm_LowScore(t *testing.T) {
	pageURL, err := url.Parse("https://example.com/contact")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`
		<html>
		<head>
			<title>Contact us</title>
			<meta name="description" content="Newsletter and general inquiries">
		</head>
		<body>
			<form action="/contact">
				<input type="text" name="fullname">
				<input type="email" name="email">
				<input type="submit" value="Send message">
			</form>
		</body>
		</html>
	`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if got.LoginScore >= 40 {
		t.Fatalf("expected low non-login score < 40, got %d (reason: %s)", got.LoginScore, got.LoginReason)
	}
	if got.HasLoginForm {
		t.Fatalf("expected HasLoginForm false for contact-style page, got true")
	}
}

func TestParseHTML_LoginReason_IncludesForgotLinkSignal(t *testing.T) {
	pageURL, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	body := strings.NewReader(`
		<html>
		<head><title>Sign in to your account</title></head>
		<body>
			<form action="/login">
				<input type="email" name="email" autocomplete="email">
				<input type="password" name="password">
				<input type="submit" value="Sign in">
			</form>
			<p><a href="https://example.com/reset">Forgot password?</a></p>
		</body>
		</html>
	`)
	got, err := ParseHTML(pageURL, body)
	if err != nil {
		t.Fatalf("ParseHTML returned error: %v", err)
	}

	if !strings.Contains(strings.ToLower(got.LoginReason), "forgot") {
		t.Fatalf("expected LoginReason to mention forgot-password link signal, got %q", got.LoginReason)
	}
}

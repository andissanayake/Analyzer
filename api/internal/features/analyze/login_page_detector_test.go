package analyze

import (
	"strings"
	"testing"
)

func TestDetectLoginPage_HighScoreLikelyLogin(t *testing.T) {
	data := LoginMetadata{
		PageTitle:       "Sign in to your account",
		PageDescription: "Use your username and password",
		Forms: []FormMetadata{
			{
				Action: "/auth/login",
				AllInputs: []InputMetadata{
					{Type: "text", Name: "username", Autocomplete: "username"},
					{Type: "password", Name: "password", Autocomplete: "current-password"},
				},
				Buttons: []ButtonMetadata{
					{Text: "Sign in"},
				},
			},
		},
		AllLinks: []LinkMetadata{
			{Href: "/forgot-password", Text: "Forgot password?"},
		},
	}

	score, reason := DetectLoginPage(data)
	if score < 70 {
		t.Fatalf("expected high login score >= 70, got %d", score)
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestDetectLoginPage_MediumScoreAuthRelated(t *testing.T) {
	data := LoginMetadata{
		PageTitle: "Account access",
		Forms: []FormMetadata{
			{
				Action: "/session",
				AllInputs: []InputMetadata{
					{Type: "password", Name: "passcode"},
				},
				Buttons: []ButtonMetadata{
					{Text: "Continue"},
				},
			},
		},
	}

	score, reason := DetectLoginPage(data)
	if score < 40 || score > 69 {
		t.Fatalf("expected medium score in [40,69], got %d", score)
	}
	if !strings.Contains(strings.ToLower(reason), "password") {
		t.Fatalf("expected reason to mention password signal, got %q", reason)
	}
}

func TestDetectLoginPage_LowScoreNonLogin(t *testing.T) {
	data := LoginMetadata{
		PageTitle:       "Contact us",
		PageDescription: "Newsletter and general inquiries",
		Forms: []FormMetadata{
			{
				Action: "/contact",
				AllInputs: []InputMetadata{
					{Type: "text", Name: "fullname"},
					{Type: "email", Name: "email"},
				},
				Buttons: []ButtonMetadata{
					{Text: "Send message"},
				},
			},
		},
	}

	score, reason := DetectLoginPage(data)
	if score >= 40 {
		t.Fatalf("expected low non-login score < 40, got %d", score)
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

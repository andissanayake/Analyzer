package analyze

import (
	"net/url"
	"strings"
)

type InputMetadata struct {
	Name         string
	ID           string
	Type         string
	Placeholder  string
	Autocomplete string
	AriaLabel    string
}

type ButtonMetadata struct {
	Name      string
	Text      string
	AriaLabel string
}

type LinkMetadata struct {
	Href      string
	Text      string
	Title     string
	AriaLabel string
}

type FormMetadata struct {
	Action    string
	AllInputs []InputMetadata
	Buttons   []ButtonMetadata
}

type LoginMetadata struct {
	Forms    []FormMetadata
	AllLinks []LinkMetadata
	PageUrl  string

	PageTitle       string
	PageDescription string
}

func DetectLoginPage(data LoginMetadata) (int, string) {
	score := 0
	reasons := make([]string, 0, 8)

	pageScore, pageReasons := scorePageText(data)
	formScore, formReasons := scoreForms(data.Forms)
	linkScore, linkReasons := scoreLinks(data.AllLinks, data.PageUrl)

	score += pageScore + formScore + linkScore
	reasons = append(reasons, pageReasons...)
	reasons = append(reasons, formReasons...)
	reasons = append(reasons, linkReasons...)

	if len(data.Forms) == 0 && scoreKeywords(data.PageTitle+" "+data.PageDescription, authKeywords) == 0 {
		score -= 10
		reasons = append(reasons, "-10 no forms and no auth-related page keywords")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "No strong login indicators were detected.")
	}

	return clampScore(score), strings.Join(reasons, "; ")
}

var (
	authKeywords       = []string{"login", "log in", "sign in", "signin", "password", "authenticate", "account access"}
	registerKeywords   = []string{"sign up", "signup", "register", "create account", "join now"}
	forgotLinkKeywords = []string{"forgot password", "reset password", "password reset", "can't sign in", "recover account"}
	formAuthPaths      = []string{"/login", "/signin", "/auth", "/session"}
	noiseKeywords      = []string{"search", "newsletter", "contact us", "subscribe"}
)

func scorePageText(data LoginMetadata) (int, []string) {
	score := 0
	reasons := make([]string, 0, 3)
	pageText := strings.TrimSpace(data.PageTitle + " " + data.PageDescription)

	if scoreKeywords(pageText, authKeywords) > 0 {
		score += 10
		reasons = append(reasons, "+10 auth keywords found in title/description")
	}
	if scoreKeywords(pageText, registerKeywords) > 0 {
		score -= 10
		reasons = append(reasons, "-10 registration keywords found in title/description")
	}
	if scoreKeywords(pageText, noiseKeywords) > 0 {
		score -= 8
		reasons = append(reasons, "-8 non-auth page intent keywords found")
	}
	return score, reasons
}

func scoreForms(forms []FormMetadata) (int, []string) {
	score := 0
	reasons := make([]string, 0, 6)
	hasPassword := false
	hasUserOrEmail := false

	for _, form := range forms {
		formHasPassword := false
		formHasUser := false
		formHasEmail := false

		if containsAnyFold(form.Action, formAuthPaths) {
			score += 6
			reasons = append(reasons, "+6 form action matches auth-like path")
		}

		for _, input := range form.AllInputs {
			if isPasswordInput(input) {
				formHasPassword = true
				hasPassword = true
			}
			if isUsernameInput(input) {
				formHasUser = true
				hasUserOrEmail = true
			}
			if isEmailInput(input) {
				formHasEmail = true
				hasUserOrEmail = true
			}
		}

		if formHasPassword {
			score += 35
			reasons = append(reasons, "+35 password input detected")
		}
		if formHasPassword && (formHasUser || formHasEmail) {
			score += 20
			reasons = append(reasons, "+20 password combined with username/email input")
		}
		buttonScore, buttonReasons := scoreButtons(form.Buttons)
		score += buttonScore
		reasons = append(reasons, buttonReasons...)
	}

	if !hasPassword && hasUserOrEmail {
		score -= 12
		reasons = append(reasons, "-12 user/email present without password input")
	}

	return score, reasons
}

func scoreButtons(buttons []ButtonMetadata) (int, []string) {
	for _, b := range buttons {
		text := strings.TrimSpace(b.Text + " " + b.AriaLabel + " " + b.Name)
		if scoreKeywords(text, []string{"login", "log in", "sign in", "continue"}) > 0 {
			return 12, []string{"+12 button text/aria indicates authentication action"}
		}
	}
	return 0, nil
}

func scoreLinks(links []LinkMetadata, pageURL string) (int, []string) {
	base, _ := url.Parse(pageURL)
	for _, l := range links {
		text := strings.TrimSpace(l.Text + " " + l.Title + " " + l.AriaLabel + " " + l.Href)
		if scoreKeywords(text, forgotLinkKeywords) > 0 || containsAnyFold(l.Href, []string{"/forgot", "/reset-password", "/password-reset"}) {
			score := 8
			reason := "+8 forgot/reset password support link detected"
			if isSameHostLink(base, l.Href) {
				score += 2
				reason += "; +2 forgot/reset link is same-domain as page URL"
			}
			return score, []string{reason}
		}
	}
	return 0, nil
}

func isSameHostLink(base *url.URL, href string) bool {
	if base == nil || base.Hostname() == "" {
		return false
	}
	ref, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return false
	}
	resolved := base.ResolveReference(ref)
	if resolved == nil || resolved.Hostname() == "" {
		return false
	}
	return strings.EqualFold(base.Hostname(), resolved.Hostname())
}

func isPasswordInput(in InputMetadata) bool {
	return strings.EqualFold(in.Type, "password") || containsAnyFold(joinInputSignals(in), []string{"current-password", "password"})
}

func isUsernameInput(in InputMetadata) bool {
	signals := joinInputSignals(in)
	return strings.EqualFold(in.Autocomplete, "username") || containsAnyFold(signals, []string{"username", "user name", "userid", "user id"})
}

func isEmailInput(in InputMetadata) bool {
	signals := joinInputSignals(in)
	return strings.EqualFold(in.Type, "email") || strings.EqualFold(in.Autocomplete, "email") || containsAnyFold(signals, []string{"email", "e-mail"})
}

func joinInputSignals(in InputMetadata) string {
	return strings.TrimSpace(in.Name + " " + in.ID + " " + in.Type + " " + in.Placeholder + " " + in.Autocomplete + " " + in.AriaLabel)
}

func scoreKeywords(text string, keywords []string) int {
	lower := strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(lower, strings.ToLower(k)) {
			return 1
		}
	}
	return 0
}

func containsAnyFold(text string, needles []string) bool {
	lower := strings.ToLower(text)
	for _, n := range needles {
		if strings.Contains(lower, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

package analyze

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(pageURL *url.URL, body io.Reader) (analysisPayload, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return analysisPayload{}, fmt.Errorf("analyze: could not parse HTML: %w", err)
	}
	return analyzeDocument(doc, pageURL)
}

func analyzeDocument(doc *html.Node, pageURL *url.URL) (analysisPayload, error) {
	var formStack []*FormMetadata
	login := LoginMetadata{PageUrl: pageURL.String()}

	payload := analysisPayload{
		HTMLVersion: "",
		PageTitle:   "",
		InternalLinks:     0,
		ExternalLinks:     0,
		InaccessibleLinks: 0,
		Headings: []headingCount{
			{Level: "h1", Count: 0},
			{Level: "h2", Count: 0},
			{Level: "h3", Count: 0},
			{Level: "h4", Count: 0},
			{Level: "h5", Count: 0},
			{Level: "h6", Count: 0},
		},
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.DoctypeNode && payload.HTMLVersion == "" {
			doctypeData := node.Data
			for _, a := range node.Attr {
				if a.Val != "" {
					doctypeData += " " + a.Val
				}
			}
			payload.HTMLVersion = htmlVersionFromDoctype(doctypeData)
		}
		if node.Type == html.ElementNode {
			switch node.Data {
			case "title":
				if t := strings.TrimSpace(textContent(node)); t != "" {
					payload.PageTitle = t
					login.PageTitle = t
				}
			case "meta":
				if strings.EqualFold(strings.TrimSpace(attrValue(node, "name")), "description") {
					if c := strings.TrimSpace(attrValue(node, "content")); c != "" {
						login.PageDescription = c
					}
				}
				if strings.EqualFold(strings.TrimSpace(attrValue(node, "property")), "og:description") && login.PageDescription == "" {
					if c := strings.TrimSpace(attrValue(node, "content")); c != "" {
						login.PageDescription = c
					}
				}
			case "a":
				isInternal, isExternal, isInaccessible := classifyAnchorLink(pageURL, node)
				switch {
				case isInternal:
					payload.InternalLinks++
				case isExternal:
					payload.ExternalLinks++
				case isInaccessible:
					payload.InaccessibleLinks++
				}
				login.AllLinks = append(login.AllLinks, linkMetadataFromAnchor(node))
			case "form":
				fm := &FormMetadata{Action: resolveFormAction(pageURL, node)}
				formStack = append(formStack, fm)
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					walk(c)
				}
				login.Forms = append(login.Forms, *formStack[len(formStack)-1])
				formStack = formStack[:len(formStack)-1]
				return
			case "input":
				inputType := strings.ToLower(strings.TrimSpace(attrValue(node, "type")))
				if inputType == "" {
					inputType = "text"
				}
				if len(formStack) > 0 {
					cur := formStack[len(formStack)-1]
					switch inputType {
					case "submit", "button", "image", "reset":
						cur.Buttons = append(cur.Buttons, buttonFromSubmitInput(node))
					default:
						cur.AllInputs = append(cur.AllInputs, inputMetadataFromNode(node))
					}
				}
			case "button":
				if len(formStack) > 0 {
					cur := formStack[len(formStack)-1]
					cur.Buttons = append(cur.Buttons, buttonElementMetadata(node))
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				level := node.Data[1] - '0'
				if level > 0 && level <= 6 {
					payload.Headings[level-1].Count++
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if payload.HTMLVersion == "" {
		payload.HTMLVersion = "HTML5"
	}
	score, reason := DetectLoginPage(login)
	payload.LoginScore = score
	payload.LoginReason = reason
	payload.HasLoginForm = score >= 70
	return payload, nil
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func htmlVersionFromDoctype(data string) string {
	d := strings.ToLower(strings.TrimSpace(data))
	switch {
	case strings.Contains(d, "html 4.01"):
		return "HTML 4.01"
	case strings.Contains(d, "xhtml 1.1"):
		return "XHTML 1.1"
	case strings.Contains(d, "xhtml 1.0"):
		return "XHTML 1.0"
	case d == "html" || strings.HasPrefix(d, "html "):
		return "HTML5"
	default:
		if d != "" {
			return strings.TrimSpace(data)
		}
		return ""
	}
}

func classifyAnchorLink(pageURL *url.URL, node *html.Node) (bool, bool, bool) {
	href := strings.TrimSpace(attrValue(node, "href"))
	if href == "" {
		return false, false, true
	}
	parsedHref, err := url.Parse(href)
	if err != nil {
		return false, false, true
	}
	absolute := pageURL.ResolveReference(parsedHref)
	if absolute == nil {
		return false, false, true
	}
	if absolute.Scheme != "http" && absolute.Scheme != "https" {
		return false, false, true
	}
	if strings.EqualFold(pageURL.Hostname(), absolute.Hostname()) {
		return true, false, false
	}
	return false, true, false
}

func attrValue(node *html.Node, key string) string {
	for _, a := range node.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func resolveFormAction(base *url.URL, form *html.Node) string {
	action := strings.TrimSpace(attrValue(form, "action"))
	if action == "" || base == nil {
		return action
	}
	u, err := url.Parse(action)
	if err != nil {
		return action
	}
	abs := base.ResolveReference(u)
	if abs == nil {
		return action
	}
	return abs.String()
}

func linkMetadataFromAnchor(n *html.Node) LinkMetadata {
	return LinkMetadata{
		Href:      attrValue(n, "href"),
		Text:      strings.TrimSpace(textContent(n)),
		Title:     attrValue(n, "title"),
		AriaLabel: attrValue(n, "aria-label"),
	}
}

func inputMetadataFromNode(n *html.Node) InputMetadata {
	return InputMetadata{
		Name:         attrValue(n, "name"),
		ID:           attrValue(n, "id"),
		Type:         attrValue(n, "type"),
		Placeholder:  attrValue(n, "placeholder"),
		Autocomplete: attrValue(n, "autocomplete"),
		AriaLabel:    attrValue(n, "aria-label"),
	}
}

func buttonFromSubmitInput(n *html.Node) ButtonMetadata {
	return ButtonMetadata{
		Name:      attrValue(n, "name"),
		Text:      attrValue(n, "value"),
		AriaLabel: attrValue(n, "aria-label"),
	}
}

func buttonElementMetadata(n *html.Node) ButtonMetadata {
	return ButtonMetadata{
		Name:      attrValue(n, "name"),
		Text:      strings.TrimSpace(textContent(n)),
		AriaLabel: attrValue(n, "aria-label"),
	}
}

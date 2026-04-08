package analyze

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const MaxHTMLBytes = 20_971_520 // 2MB

func ParseHTML(pageURL *url.URL, body io.Reader) (analysisPayload, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return analysisPayload{}, fmt.Errorf("analyze: could not parse HTML: %w", err)
	}
	return analyzeDocument(doc, pageURL)
}

func analyzeDocument(doc *html.Node, pageURL *url.URL) (analysisPayload, error) {
	payload := analysisPayload{
		InternalLinks:     0,
		ExternalLinks:     0,
		InaccessibleLinks: 0,
		HasLoginForm:      false,
		Headings: []headingCount{
			{Level: "h1", Count: 0},
			{Level: "h2", Count: 0},
			{Level: "h3", Count: 0},
			{Level: "h4", Count: 0},
			{Level: "h5", Count: 0},
			{Level: "h6", Count: 0},
		},
	}
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			payload.HTMLVersion = htmlVersionFromDoctype(c.Data)
			break
		}
	}
	if payload.HTMLVersion == "" {
		payload.HTMLVersion = "HTML5"
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "title":
				if t := strings.TrimSpace(textContent(node)); t != "" {
					payload.PageTitle = t
				}
			case "a":
				payload.InternalLinks++
			case "img":
				payload.InaccessibleLinks++
			case "form":
				payload.HasLoginForm = true
			case "input":
				for _, a := range node.Attr {
					if a.Key == "type" && strings.EqualFold(a.Val, "password") {
						payload.HasLoginForm = true
						break
					}
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

package analyze

import "errors"

var errURLNotReachable = errors.New("url is not reachable")

func analyzeURL(url string) (analysisResult, error) {
	if url == "" {
		return analysisResult{}, errURLNotReachable
	}

	return analysisResult{
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
	}, nil
}

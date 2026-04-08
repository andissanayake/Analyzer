package analyze

type request struct {
	URL string `json:"url"`
}

type headingCount struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

type analysisPayload struct {
	HTMLVersion       string         `json:"htmlVersion"`
	PageTitle         string         `json:"pageTitle"`
	Headings          []headingCount `json:"headings"`
	InternalLinks     int            `json:"internalLinks"`
	ExternalLinks     int            `json:"externalLinks"`
	InaccessibleLinks int            `json:"inaccessibleLinks"`
	HasLoginForm      bool           `json:"hasLoginForm"`
}

type analysisResult struct {
	StatusCode int              `json:"statusCode"`
	Message    string           `json:"message"`
	Body       *analysisPayload `json:"body,omitempty"`
}

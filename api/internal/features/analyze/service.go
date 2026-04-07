package analyze

import (
	"errors"
	"net/http"
)

var errURLNotReachable = errors.New("url is not reachable")

func analyzeURL(url string) (analysisResult, error) {
	if url == "" {
		return analysisResult{StatusCode: http.StatusBadRequest, Message: "url is required"}, errURLNotReachable
	}

	return analysisResult{StatusCode: http.StatusOK, Message: "Analysis complete.", Body: analysisPayload{
			InternalLinks:     0,
			ExternalLinks:     0,
			InaccessibleLinks: 0,
			HasLoginForm:      false,
		},
	}, nil
}

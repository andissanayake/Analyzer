package analyze
import (
	"errors"
	"testing"
)
func TestAnalyzeURL_EmptyURL_ReturnsReachableError(t *testing.T) {
	_, err := analyzeURL("")
	if !errors.Is(err, errURLNotReachable) {
		t.Fatalf("expected errURLNotReachable, got %v", err)
	}
}
package notifications

import (
	"strings"
	"testing"
)

func TestGeneratedOperationsMatchPublicAllowlist(t *testing.T) {
	t.Parallel()
	if len(allOperations) != 29 {
		t.Fatalf("operations=%d, want 29", len(allOperations))
	}
	for _, operation := range allOperations {
		if !strings.HasPrefix(operation.Path, "/v1/notifications/") {
			t.Fatalf("non-public operation generated: %s", operation.Path)
		}
		if strings.Contains(operation.Path, "/inbox") || strings.Contains(operation.Path, "/devices") || strings.Contains(operation.Path, "/preferences") {
			t.Fatalf("partner-private operation generated: %s", operation.Path)
		}
	}
}

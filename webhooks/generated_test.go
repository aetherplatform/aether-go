package webhooks

import (
	"strings"
	"testing"
)

func TestGeneratedOperationsMatchPublicAllowlist(t *testing.T) {
	t.Parallel()
	if len(allOperations) != 19 {
		t.Fatalf("operations=%d, want 19", len(allOperations))
	}
	for _, operation := range allOperations {
		if !strings.HasPrefix(operation.Path, "/v1/webhooks/") {
			t.Fatalf("non-public operation generated: %s", operation.Path)
		}
	}
}

package notifications

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNestedTemplatePreviewDecodesNullableChannels(t *testing.T) {
	t.Parallel()
	var preview TemplatePreview
	if err := json.Unmarshal([]byte(`{"slug":"welcome","locale":"en","rendered":{"email_subject":"Hello Ada","sms_body":null,"source":"db"}}`), &preview); err != nil {
		t.Fatal(err)
	}
	if preview.Slug != "welcome" || preview.Rendered.EmailSubject == nil || *preview.Rendered.EmailSubject != "Hello Ada" || preview.Rendered.SmsBody != nil {
		t.Fatalf("unexpected preview: %+v", preview)
	}
}

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

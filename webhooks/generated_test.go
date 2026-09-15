package webhooks

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aetherplatform/aether-go"
)

type staticTokenProvider struct{}

func (staticTokenProvider) Token(context.Context) (aether.AccessToken, error) {
	return aether.AccessToken{Value: "sdk-token", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestQueuedReplayReturnsAttemptWithoutRepeatingRequest(t *testing.T) {
	t.Parallel()
	requests := 0
	client, err := NewClient(aether.Config{
		BaseURL: "https://api.example", TokenProvider: staticTokenProvider{},
		HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			requests++
			if r.Method != "POST" || r.URL.Path != "/v1/webhooks/deliveries/delivery_1/replay" {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			return &http.Response{StatusCode: 202, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":"queued","attempt_id":"attempt_1"}`))}, nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ReplayWebhookDelivery(context.Background(), "delivery_1")
	if err != nil {
		t.Fatal(err)
	}
	if (*result)["attempt_id"] != "attempt_1" || requests != 1 {
		t.Fatalf("result=%v requests=%d", result, requests)
	}
}

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

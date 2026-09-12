package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aetherplatform/aether-go"
)

type retryTestTokenProvider struct{}

func (retryTestTokenProvider) Token(context.Context) (aether.AccessToken, error) {
	return aether.AccessToken{Value: "test-token", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func TestPublishWebhookEventRetriesOnlyWithUsableBodyKey(t *testing.T) {
	t.Parallel()
	empty, blank, key := "", " \t", "event-123"
	for _, test := range []struct {
		name string
		key  *string
		want int64
	}{
		{"missing", nil, 1}, {"empty", &empty, 1}, {"blank", &blank, 1}, {"keyed", &key, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				var body PublishEventRequest
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if (body.IdempotencyKey == nil) != (test.key == nil) || (test.key != nil && body.IdempotencyKey != nil && *body.IdempotencyKey != *test.key) {
					t.Error("idempotency key changed on the wire")
				}
				writer.Header().Set("Retry-After", "0")
				if requests.Add(1) == 1 {
					writer.WriteHeader(http.StatusServiceUnavailable)
				} else {
					writer.WriteHeader(http.StatusAccepted)
				}
				_, _ = writer.Write([]byte(`{}`))
			}))
			defer server.Close()
			client, err := NewClient(aether.Config{BaseURL: server.URL, TokenProvider: retryTestTokenProvider{}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.PublishWebhookEvent(context.Background(), PublishEventRequest{Type: "example", Data: JsonObject{}, IdempotencyKey: test.key})
			if (err == nil) != (test.want == 2) {
				t.Fatalf("error=%v, want %d attempts", err, test.want)
			}
			if got := requests.Load(); got != test.want {
				t.Fatalf("requests=%d, want %d", got, test.want)
			}
		})
	}
}

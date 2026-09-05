package events

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

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestGeneratedOperationsMatchPublicAllowlist(t *testing.T) {
	t.Parallel()
	if len(allOperations) != 3 {
		t.Fatalf("operations=%d, want 3", len(allOperations))
	}
	for _, operation := range allOperations {
		if !strings.HasPrefix(operation.Path, "/v1/events/") || strings.Contains(operation.Path, "consumer-groups") {
			t.Fatalf("non-public operation generated: %s", operation.Path)
		}
	}
}

func TestEventSchemaPathEncoding(t *testing.T) {
	t.Parallel()
	var path string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		path = request.URL.EscapedPath()
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"type":"storage.object.created","version":2,"schema":{}}}`))}, nil
	})}
	client, err := NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: staticTokenProvider{}, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetEventSchema(context.Background(), "storage.object.created", 2); err != nil {
		t.Fatal(err)
	}
	if path != "/v1/events/schemas/storage.object.created/2" {
		t.Fatalf("path=%q", path)
	}
}

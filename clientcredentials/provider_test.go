//go:build !js

package clientcredentials

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestRetryConfiguration(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		max      int
		disabled bool
		want     int64
	}{
		{"default", 0, false, 2}, {"explicit", 2, false, 3},
		{"disabled", 0, true, 1}, {"disabled overrides positive", 3, true, 1},
		{"negative remains invalid", -1, true, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int64
			provider, err := New(Config{
				TokenURL: "https://auth.example/oauth/token", ClientID: "client", ClientSecret: "secret",
				Audience: "aether-storage", Capabilities: []string{"storage:*"},
				MaxRetries: test.max, DisableRetries: test.disabled,
				HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					calls.Add(1)
					result := response(http.StatusServiceUnavailable, `{}`)
					result.Header.Set("Retry-After", "0")
					return result, nil
				})},
			})
			if test.want == 0 {
				if err == nil {
					t.Fatal("negative retries accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.Token(context.Background())
			if err == nil || calls.Load() != test.want {
				t.Fatalf("error=%v calls=%d, want %d", err, calls.Load(), test.want)
			}
		})
	}
}

func TestProviderCoalescesAndCachesTokenRequests(t *testing.T) {
	t.Parallel()
	var requests atomic.Int64
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if clientID, secret, ok := request.BasicAuth(); !ok || clientID != "client" || secret != "secret" {
			t.Errorf("unexpected basic authentication")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if request.Form.Get("scope") != "storage:objects/*:read" {
			t.Errorf("unexpected scope: %q", request.Form.Get("scope"))
		}
		time.Sleep(20 * time.Millisecond)
		return response(http.StatusOK, `{"access_token":"token","token_type":"Bearer","expires_in":600}`), nil
	})}

	provider, err := New(Config{TokenURL: "https://auth.useather.test/oauth/token", ClientID: "client", ClientSecret: "secret", Audience: "aether-storage", Capabilities: []string{"storage:objects/*:read"}, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for range 12 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			token, tokenErr := provider.Token(context.Background())
			if tokenErr != nil || token.Value != "token" {
				t.Errorf("token=%q error=%v", token.Value, tokenErr)
			}
		}()
	}
	wait.Wait()
	if requests.Load() != 1 {
		t.Fatalf("requests=%d, want 1", requests.Load())
	}
	if _, err := provider.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("cached request count=%d, want 1", requests.Load())
	}
}

func TestProviderBoundsInvalidClientAndRedactsSecret(t *testing.T) {
	t.Parallel()
	var requests atomic.Int64
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		requests.Add(1)
		return response(http.StatusUnauthorized, `{"error":{"code":"invalid_client","message":"bad secret-value"}}`), nil
	})}

	provider, err := New(Config{TokenURL: "https://auth.useather.test/oauth/token", ClientID: "client", ClientSecret: "secret-value", Audience: "aether-storage", Capabilities: []string{"storage:*"}, MaxRetries: 3, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Token(context.Background())
	if err == nil {
		t.Fatal("expected invalid client error")
	}
	if requests.Load() != 1 {
		t.Fatalf("requests=%d, want 1", requests.Load())
	}
	if strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("secret leaked in error: %v", err)
	}
}

func TestProviderBoundsTransientFailure(t *testing.T) {
	t.Parallel()
	var requests atomic.Int64
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		requests.Add(1)
		return response(http.StatusServiceUnavailable, `{}`), nil
	})}

	provider, err := New(Config{TokenURL: "https://auth.useather.test/oauth/token", ClientID: "client", ClientSecret: "secret", Audience: "aether-storage", Capabilities: []string{"storage:*"}, MaxRetries: 1, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Token(context.Background())
	if err == nil {
		t.Fatal("expected transient error")
	}
	if requests.Load() != 2 {
		t.Fatalf("requests=%d, want 2", requests.Load())
	}
}

func TestProviderCancelsAcquisitionWhenAllWaitersCancel(t *testing.T) {
	t.Parallel()
	requestCancelled := make(chan struct{})
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		close(requestCancelled)
		return nil, request.Context().Err()
	})}
	provider, err := New(Config{
		TokenURL:     "https://auth.useather.test/oauth/token",
		ClientID:     "client",
		ClientSecret: "secret",
		Audience:     "aether-storage",
		Capabilities: []string{"storage:*"},
		HTTPClient:   httpClient,
		Timeout:      time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, tokenErr := provider.Token(ctx)
		result <- tokenErr
	}()
	cancel()
	if err := <-result; err == nil {
		t.Fatal("expected cancellation error")
	}
	select {
	case <-requestCancelled:
	case <-time.After(time.Second):
		t.Fatal("in-flight acquisition was not cancelled")
	}
}

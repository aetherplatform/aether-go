package identity

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aetherplatform/aether-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type rotatingTokenProvider struct {
	mu          sync.Mutex
	tokens      []string
	index       int
	invalidated int
}

func (provider *rotatingTokenProvider) Token(context.Context) (aether.AccessToken, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return aether.AccessToken{Value: provider.tokens[provider.index], ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (provider *rotatingTokenProvider) Invalidate() {
	provider.mu.Lock()
	provider.invalidated++
	if provider.index < len(provider.tokens)-1 {
		provider.index++
	}
	provider.mu.Unlock()
}

func TestAuthorizationURLBuildsMandatoryPKCERequest(t *testing.T) {
	t.Parallel()
	client, err := NewClient(Config{BaseURL: "https://auth.useather.test"})
	if err != nil {
		t.Fatal(err)
	}
	target, err := client.AuthorizationURL(AuthorizationRequest{
		ClientID:      "client_1",
		RedirectURI:   "https://app.example/callback",
		Scope:         "openid profile",
		State:         "state_1",
		Nonce:         "nonce_1",
		CodeChallenge: strings.Repeat("a", 43),
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if parsed.Path != "/oauth/authorize" || query.Get("response_type") != "code" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("unexpected authorization URL: %s", target)
	}
	if query.Has("principal_id") {
		t.Fatal("authorization URL exposed principal_id")
	}
}

func TestDiscoveryRetriesTransientFailure(t *testing.T) {
	t.Parallel()
	requests := 0
	client, err := NewClient(Config{BaseURL: "https://auth.useather.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Authorization") != "" {
			t.Fatal("discovery request included authorization")
		}
		if requests == 1 {
			return identityResponse(http.StatusServiceUnavailable, `{"error":{"code":"temporarily_unavailable","message":"retry"}}`), nil
		}
		return identityResponse(http.StatusOK, `{"issuer":"https://auth.useather.test","authorization_endpoint":"https://auth.useather.test/oauth/authorize","token_endpoint":"https://auth.useather.test/oauth/token","jwks_uri":"https://auth.useather.test/.well-known/jwks.json","response_types_supported":["code"],"grant_types_supported":["authorization_code"],"code_challenge_methods_supported":["S256"],"id_token_signing_alg_values_supported":["EdDSA","RS256"]}`), nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	configuration, err := client.GetOpenIDConfiguration(context.Background(), aether.WithCorrelationID("corr_1"))
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || configuration.Issuer != "https://auth.useather.test" {
		t.Fatalf("requests=%d issuer=%q", requests, configuration.Issuer)
	}
}

func TestUserInfoRefreshesBearerOnce(t *testing.T) {
	t.Parallel()
	provider := &rotatingTokenProvider{tokens: []string{"stale", "fresh"}}
	requests := 0
	client, err := NewClient(Config{BaseURL: "https://auth.useather.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Authorization") == "Bearer stale" {
			return identityResponse(http.StatusUnauthorized, `{}`), nil
		}
		return identityResponse(http.StatusOK, `{"sub":"pairwise_subject"}`), nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	info, err := client.GetUserInfo(context.Background(), provider)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || provider.invalidated != 1 || info.Subject() != "pairwise_subject" {
		t.Fatalf("requests=%d invalidations=%d subject=%q", requests, provider.invalidated, info.Subject())
	}
}

func TestUserInfoDoesNotLoopAfterFreshTokenIsRejected(t *testing.T) {
	t.Parallel()
	provider := &rotatingTokenProvider{tokens: []string{"stale", "fresh"}}
	requests := 0
	client, err := NewClient(Config{BaseURL: "https://auth.useather.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return identityResponse(http.StatusUnauthorized, `{}`), nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetUserInfo(context.Background(), provider)
	var requestErr *aether.Error
	if !errors.As(err, &requestErr) || requestErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("error=%v", err)
	}
	if requests != 2 || provider.invalidated != 1 {
		t.Fatalf("requests=%d invalidations=%d", requests, provider.invalidated)
	}
}

func identityResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

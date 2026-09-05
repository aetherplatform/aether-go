//go:build !js

package identity

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/aetherplatform/aether-go"
)

func TestConfidentialClientUsesBasicAuthAndFormEncoding(t *testing.T) {
	t.Parallel()
	var authorization string
	var form url.Values
	client, err := NewConfidentialClient(ConfidentialConfig{
		Config: Config{BaseURL: "https://auth.useather.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			authorization = request.Header.Get("Authorization")
			body, _ := io.ReadAll(request.Body)
			form, _ = url.ParseQuery(string(body))
			return identityResponse(http.StatusOK, `{"access_token":"access","token_type":"Bearer","expires_in":600}`), nil
		})}},
		ClientID: "client_1", ClientSecret: "secret_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ExchangeOAuthToken(context.Background(), TokenRequest{GrantType: GrantClientCredentials, Audience: "aether-events", Scope: "events:catalog/*:read"})
	if err != nil {
		t.Fatal(err)
	}
	wantAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte("client_1:secret_1"))
	if authorization != wantAuthorization || form.Get("grant_type") != "client_credentials" || form.Get("scope") != "events:catalog/*:read" {
		t.Fatalf("authorization=%q form=%v", authorization, form)
	}
}

func TestConfidentialClientDoesNotRetryOrExposeInvalidSecret(t *testing.T) {
	t.Parallel()
	requests := 0
	client, err := NewConfidentialClient(ConfidentialConfig{
		Config: Config{BaseURL: "https://auth.useather.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requests++
			return identityResponse(http.StatusUnauthorized, `{"error":{"code":"invalid_client","message":"secret_1 rejected","details":{"secret":"secret_1"}}}`), nil
		})}},
		ClientID: "client_1", ClientSecret: "secret_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ExchangeOAuthToken(context.Background(), TokenRequest{GrantType: GrantClientCredentials, Audience: "aether-events", Scope: "events:catalog/*:read"})
	var requestErr *aether.Error
	if !errors.As(err, &requestErr) {
		t.Fatalf("error=%v", err)
	}
	if requests != 1 || strings.Contains(requestErr.Message, "secret_1") || requestErr.Details != nil {
		t.Fatalf("requests=%d error=%+v", requests, requestErr)
	}
}

func TestOAuthGrantForms(t *testing.T) {
	t.Parallel()
	verifier := strings.Repeat("v", 43)
	tests := []struct {
		name    string
		request TokenRequest
		want    url.Values
	}{
		{
			name:    "authorization code",
			request: TokenRequest{GrantType: GrantAuthorizationCode, Code: "code_1", RedirectURI: "https://app.example/callback", CodeVerifier: verifier},
			want:    url.Values{"grant_type": {"authorization_code"}, "code": {"code_1"}, "redirect_uri": {"https://app.example/callback"}, "code_verifier": {verifier}},
		},
		{
			name:    "refresh token",
			request: TokenRequest{GrantType: GrantRefreshToken, RefreshToken: "refresh_1"},
			want:    url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"refresh_1"}},
		},
		{
			name:    "client credentials",
			request: TokenRequest{GrantType: GrantClientCredentials, Audience: "aether-events", Scope: "events:catalog/*:read"},
			want:    url.Values{"grant_type": {"client_credentials"}, "audience": {"aether-events"}, "scope": {"events:catalog/*:read"}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			form, _, err := tokenRequestForm(test.request)
			if err != nil {
				t.Fatal(err)
			}
			if form.Encode() != test.want.Encode() {
				t.Fatalf("form=%q, want %q", form.Encode(), test.want.Encode())
			}
		})
	}
}

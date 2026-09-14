package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aetherplatform/aether-go"
)

func startRequest() PasswordlessStartRequest {
	return PasswordlessStartRequest{Identifier: "person@example.test", Channel: ChannelEmail, RedirectURI: "https://app.example/callback", CodeChallenge: strings.Repeat("a", 43)}
}
func verifyRequest() PasswordlessVerifyRequest {
	return PasswordlessVerifyRequest{Transaction: "transaction-secret", ChallengeID: "challenge", Identifier: "person@example.test", Channel: ChannelEmail, Code: "123456", CodeVerifier: strings.Repeat("v", 43)}
}
func completeRequest() PasswordlessCompleteRequest {
	return PasswordlessCompleteRequest{Continuation: "continuation-secret", CodeVerifier: strings.Repeat("v", 43), Decision: DecisionDeny}
}
func publicTestClient(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()
	client, err := NewClient(Config{BaseURL: "https://auth.example", ClientID: "public-client", MaxRetries: 5, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestPasswordlessStartSendsPublicJSONAndPreflightHint(t *testing.T) {
	t.Parallel()
	client := publicTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/passwordless/start" || req.URL.Query().Get("client_id") != "public-client" || req.Header.Get("Authorization") != "" || req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("wrong public request metadata")
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["client_id"] != "public-client" || body["code_challenge_method"] != "S256" || body["channel"] != "email" {
			t.Fatalf("invalid request shape")
		}
		if _, exists := body["client_secret"]; exists {
			t.Fatal("public client transmitted a secret")
		}
		return identityResponse(202, `{"transaction":"transaction","challenge_id":"challenge","expires_in":300,"resend_after":60,"transaction_expires_in":600}`), nil
	}))
	response, err := client.StartPasswordless(context.Background(), startRequest())
	if err != nil || response.ResendAfter != 60 {
		t.Fatalf("start: %v", err)
	}
}

func TestPasswordlessStartAcceptsConfiguredResendIntervals(t *testing.T) {
	t.Parallel()
	for _, seconds := range []string{"1", "30", "60", "300", "0", "-1", "301", "1.5", `"30"`, "null"} {
		t.Run(seconds, func(t *testing.T) {
			calls := 0
			client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return identityResponse(202, fmt.Sprintf(`{"transaction":"transaction","challenge_id":"challenge","expires_in":300,"resend_after":%s,"transaction_expires_in":600}`, seconds)), nil
			}))
			response, err := client.StartPasswordless(context.Background(), startRequest())
			valid := seconds == "1" || seconds == "30" || seconds == "60" || seconds == "300"
			if valid {
				if err != nil || fmt.Sprint(response.ResendAfter) != seconds {
					t.Fatalf("configured resend interval was not preserved: %v", err)
				}
			} else if err == nil {
				t.Fatal("invalid resend interval was accepted")
			}
			if calls != 1 {
				t.Fatalf("sending was retried: %d calls", calls)
			}
		})
	}
}

func TestPasswordlessOutcomeTypesAndMalformedResponses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		body  string
		valid bool
		kind  string
	}{
		{`{"status":"authorized","code":"code","state":"state"}`, true, "authorized"},
		{`{"status":"hosted_completion_required","continuation_url":"https://auth.example/oauth/passwordless/continue?handoff=secret","expires_in":300}`, true, "hosted"},
		{`{"status":"custom_completion_required","continuation":"continuation","expires_in":300,"requirements":{"registration":true,"consent":true,"scopes":["openid"],"aether_terms_version":"1","aether_privacy_version":"1"}}`, true, "custom"},
		{`{"status":"authorized","code":"code"}`, false, ""},
		{`{"status":"authorized","code":"","state":null}`, false, ""},
		{`{"status":"unknown","code":"code","state":null}`, false, ""},
		{`{"status":"denied","error":"access_denied","state":null}`, false, ""},
		{`{"status":"custom_completion_required","continuation":"continuation","expires_in":300,"requirements":{}}`, false, ""},
		{`{"status":"hosted_completion_required","continuation_url":"javascript:secret","expires_in":300}`, false, ""},
	}
	for _, test := range tests {
		t.Run(test.kind+test.body[:20], func(t *testing.T) {
			client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) { return identityResponse(200, test.body), nil }))
			result, err := client.VerifyPasswordless(context.Background(), verifyRequest())
			if !test.valid {
				if err == nil || result != nil {
					t.Fatal("accepted invalid outcome")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			switch result.(type) {
			case *PasswordlessAuthorized:
				if test.kind != "authorized" {
					t.Fatal("wrong type")
				}
			case *PasswordlessHostedCompletionRequired:
				if test.kind != "hosted" {
					t.Fatal("wrong type")
				}
			case *PasswordlessCustomCompletionRequired:
				if test.kind != "custom" {
					t.Fatal("wrong type")
				}
			default:
				t.Fatal("unknown type")
			}
		})
	}
}

func TestPasswordlessCompletionDenialAndPublicOAuth(t *testing.T) {
	t.Parallel()
	client := publicTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "" {
			t.Fatal("unexpected credentials")
		}
		if req.URL.Path == "/v1/passwordless/complete" {
			return identityResponse(200, `{"status":"denied","error":"access_denied","state":"state"}`), nil
		}
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if req.Form.Get("client_id") != "public-client" || req.Form.Get("client_secret") != "" {
			t.Fatal("invalid public OAuth form")
		}
		return identityResponse(200, `{"access_token":"access","token_type":"Bearer","expires_in":3600}`), nil
	}))
	outcome, err := client.CompletePasswordless(context.Background(), completeRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := outcome.(*PasswordlessDenied); !ok {
		t.Fatal("expected denied outcome")
	}
	for _, request := range []TokenRequest{{GrantType: GrantAuthorizationCode, Code: "code", RedirectURI: "https://app.example/callback", CodeVerifier: strings.Repeat("v", 43)}, {GrantType: GrantRefreshToken, RefreshToken: "refresh"}} {
		if _, err := client.ExchangeOAuthToken(context.Background(), request); err != nil {
			t.Fatal(err)
		}
	}
	if err := client.RevokeOAuthToken(context.Background(), TokenHandleRequest{Token: "refresh"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ExchangeOAuthToken(context.Background(), TokenRequest{GrantType: GrantClientCredentials, Audience: "aether-events", Scope: "read"}); err == nil {
		t.Fatal("public machine grant accepted")
	}
}

func TestAuthenticationMutationsNeverRetryAndSuppressAllErrorChannels(t *testing.T) {
	t.Parallel()
	calls := []func(*Client) error{
		func(c *Client) error { _, err := c.StartPasswordless(context.Background(), startRequest()); return err },
		func(c *Client) error {
			_, err := c.VerifyPasswordless(context.Background(), verifyRequest())
			return err
		},
		func(c *Client) error {
			_, err := c.CompletePasswordless(context.Background(), completeRequest())
			return err
		},
		func(c *Client) error {
			_, err := c.ExchangeOAuthToken(context.Background(), TokenRequest{GrantType: GrantAuthorizationCode, Code: "code-secret", RedirectURI: "https://app.example/callback", CodeVerifier: strings.Repeat("v", 43)})
			return err
		},
		func(c *Client) error {
			_, err := c.ExchangeOAuthToken(context.Background(), TokenRequest{GrantType: GrantRefreshToken, RefreshToken: "refresh-secret"})
			return err
		},
		func(c *Client) error {
			return c.RevokeOAuthToken(context.Background(), TokenHandleRequest{Token: "refresh-secret"})
		},
	}
	for index, call := range calls {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			count := 0
			client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
				count++
				response := identityResponse(429, `{"error":{"code":"RATE_LIMITED","message":"secret_123456","request_id":"secret_123456","details":{"token":"secret_123456"}}}`)
				response.Header.Set("x-request-id", "secret_123456")
				response.Header.Set("quota-reset", "secret_123456")
				response.Header.Set("retry-after", "60")
				return response, nil
			}))
			err := call(client)
			var requestErr *aether.Error
			if count != 1 || !errors.As(err, &requestErr) || requestErr.Code != "RATE_LIMITED" || requestErr.Retry.RetryAfterSeconds != 60 {
				t.Fatalf("request count=%d err=%v", count, err)
			}
			if strings.Contains(fmt.Sprintf("%+v", requestErr), "secret") || requestErr.Cause != nil || requestErr.Details != nil {
				t.Fatal("sensitive error leaked")
			}
		})
	}
}

func TestTransportFailureAndRedirectCannotExposeCredentials(t *testing.T) {
	t.Parallel()
	client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, fmt.Errorf("code-secret and 123456") }))
	_, err := client.VerifyPasswordless(context.Background(), verifyRequest())
	var requestErr *aether.Error
	if !errors.As(err, &requestErr) || requestErr.Cause != nil || strings.Contains(fmt.Sprintf("%+v", err), "123456") {
		t.Fatal("transport cause leaked")
	}
	reached := false
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	redirected, _ := NewClient(Config{BaseURL: source.URL, ClientID: "public"})
	if _, err := redirected.StartPasswordless(context.Background(), startRequest()); err == nil || reached {
		t.Fatal("authentication redirect followed")
	}
}

func TestInvalidInputAndWrongSuccessStatusFailBeforeUse(t *testing.T) {
	t.Parallel()
	calls := 0
	client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return identityResponse(200, `{"transaction":"secret","challenge_id":"id","expires_in":300,"resend_after":60,"transaction_expires_in":600}`), nil
	}))
	request := startRequest()
	request.CodeChallenge = "short"
	if _, err := client.StartPasswordless(context.Background(), request); err == nil || calls != 0 {
		t.Fatal("invalid input sent")
	}
	if response, err := client.StartPasswordless(context.Background(), startRequest()); err == nil || response != nil {
		t.Fatal("undocumented status accepted")
	}
	other := "other-client"
	request = startRequest()
	request.ClientID = &other
	if _, err := client.StartPasswordless(context.Background(), request); err == nil || calls != 1 {
		t.Fatal("foreign client sent")
	}
}

func TestPartialReadFailureIsSanitized(t *testing.T) {
	t.Parallel()
	client := publicTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: failingBody{}}, nil
	}))
	_, err := client.VerifyPasswordless(context.Background(), verifyRequest())
	if strings.Contains(fmt.Sprintf("%+v", err), "secret") {
		t.Fatal("read failure leaked")
	}
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, fmt.Errorf("secret token read failure") }
func (failingBody) Close() error             { return nil }

var _ io.ReadCloser = failingBody{}

func TestPublicUserInfoIncludesBrowserPreflightClientHint(t *testing.T) {
	t.Parallel()
	client := publicTestClient(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/oauth/userinfo" || request.URL.Query().Get("client_id") != "public-client" {
			t.Fatal("userinfo must identify the registered client for browser preflight")
		}
		return identityResponse(http.StatusOK, `{"sub":"pairwise_subject"}`), nil
	}))
	_, err := client.GetUserInfo(context.Background(), &rotatingTokenProvider{tokens: []string{"access-token"}})
	if err != nil {
		t.Fatal(err)
	}
}

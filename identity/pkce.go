package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type PKCE struct {
	Verifier        string
	Challenge       string
	ChallengeMethod string
}

// GeneratePKCE creates a fresh S256 proof. Keep the verifier in application-owned
// storage until code exchange; never send it to StartPasswordless or log it.
func GeneratePKCE() (PKCE, error) {
	verifier, err := GenerateState()
	if err != nil {
		return PKCE{}, err
	}
	digest := sha256.Sum256([]byte(verifier))
	return PKCE{Verifier: verifier, Challenge: base64.RawURLEncoding.EncodeToString(digest[:]), ChallengeMethod: "S256"}, nil
}

// GenerateState returns a cryptographically random callback binding.
func GenerateState() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("identity: secure random generation failed")
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}

type CallbackResult struct{ Code, State string }

type OAuthCallbackError struct{ Code string }

func (err *OAuthCallbackError) Error() string {
	return "identity: OAuth authorization failed (" + err.Code + ")"
}

// ValidateOAuthCallback validates the exact registered callback and original
// state before exposing a code. It rejects duplicates, fragments and mixed
// success/error responses. OAuth error descriptions are intentionally discarded.
func ValidateOAuthCallback(callbackURL, redirectURI, expectedState string) (*CallbackResult, error) {
	invalid := func() (*CallbackResult, error) { return nil, fmt.Errorf("identity: invalid OAuth callback") }
	expected, err := validateRedirectURI(redirectURI)
	if err != nil || expectedState == "" {
		return invalid()
	}
	callback, err := url.Parse(callbackURL)
	if err != nil || strings.Contains(callbackURL, "#") || callback.Fragment != "" || callback.User != nil {
		return invalid()
	}
	rawQuery := callback.RawQuery
	callback.RawQuery, callback.ForceQuery = "", false
	if callback.String() != expected.String() {
		return invalid()
	}
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return invalid()
	}
	for key, values := range query {
		if len(values) != 1 {
			return invalid()
		}
		switch key {
		case "code", "state", "error", "error_description", "error_uri":
		default:
			return invalid()
		}
	}
	state := query.Get("state")
	if subtle.ConstantTimeCompare([]byte(state), []byte(expectedState)) != 1 {
		return invalid()
	}
	if code := query.Get("error"); code != "" {
		if _, present := query["code"]; present {
			return invalid()
		}
		switch code {
		case "access_denied", "invalid_request", "unauthorized_client", "unsupported_response_type", "invalid_scope", "server_error", "temporarily_unavailable":
			return nil, &OAuthCallbackError{Code: code}
		default:
			return invalid()
		}
	}
	if _, present := query["error"]; present {
		return invalid()
	}
	if _, present := query["error_description"]; present {
		return invalid()
	}
	if _, present := query["error_uri"]; present {
		return invalid()
	}
	if query.Get("code") == "" {
		return invalid()
	}
	return &CallbackResult{Code: query.Get("code"), State: state}, nil
}

func validateRedirectURI(value string) (*url.URL, error) {
	target, err := url.Parse(value)
	if err != nil || strings.Contains(value, "#") || target.Host == "" || target.User != nil || target.RawQuery != "" || target.ForceQuery || target.Fragment != "" {
		return nil, fmt.Errorf("identity: redirect URI must be an exact absolute callback without query or fragment")
	}
	if target.Scheme == "https" {
		return target, nil
	}
	// Only explicitly registered development callbacks are accepted by Aether.
	if target.Scheme == "http" {
		switch target.Hostname() {
		case "127.0.0.1", "::1", "localhost":
			return target, nil
		}
	}
	return nil, fmt.Errorf("identity: redirect URI requires HTTPS or development loopback HTTP")
}

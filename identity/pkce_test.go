package identity

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
)

func TestPKCEUsesFreshS256Proof(t *testing.T) {
	t.Parallel()
	first, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(first.Verifier))
	if first.Verifier == second.Verifier || !codeVerifierPattern.MatchString(first.Verifier) || first.Challenge != base64.RawURLEncoding.EncodeToString(digest[:]) || first.ChallengeMethod != "S256" {
		t.Fatal("invalid PKCE")
	}
}
func TestStrictCallbackValidation(t *testing.T) {
	t.Parallel()
	redirect := "https://app.example/callback"
	valid, err := ValidateOAuthCallback(redirect+"?code=code&state=state", redirect, "state")
	if err != nil || valid.Code != "code" {
		t.Fatal(err)
	}
	for _, callback := range []string{
		"https://evil.example/callback?code=code&state=state",
		redirect + "?code=code&state=wrong", redirect + "?code=code&state=state&state=state",
		redirect + "?code=one&code=two&state=state", redirect + "?code=code&state=state#error",
		redirect + "?code=code&state=state&error=access_denied", redirect + "?code=code&state=state&unexpected=value",
		redirect + "?state=state", redirect + "?code=code&state=state&error=",
		"https://app.example/other?code=code&state=state",
	} {
		if _, err := ValidateOAuthCallback(callback, redirect, "state"); err == nil {
			t.Fatalf("accepted invalid callback %q", callback)
		}
	}
	_, err = ValidateOAuthCallback(redirect+"?error=access_denied&error_description=secret&state=state", redirect, "state")
	var denial *OAuthCallbackError
	if !errors.As(err, &denial) || denial.Code != "access_denied" {
		t.Fatal("denial lost")
	}
}

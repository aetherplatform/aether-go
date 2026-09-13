package identity

import (
	"context"
	"fmt"
	"github.com/aetherplatform/aether-go"
	"net/url"
	"strings"
)

// ExchangeOAuthToken exchanges authorization codes or rotates refresh tokens
// for the registered public ClientID. Machine grants require ConfidentialClient.
func (client *Client) ExchangeOAuthToken(ctx context.Context, request TokenRequest, options ...aether.RequestOption) (*TokenResponse, error) {
	if client.clientID == "" {
		return nil, fmt.Errorf("identity: public client ID is required")
	}
	if request.GrantType != GrantAuthorizationCode && request.GrantType != GrantRefreshToken {
		return nil, fmt.Errorf("identity: public clients cannot use this OAuth grant")
	}
	form, sensitive, err := tokenRequestForm(request)
	if err != nil {
		return nil, err
	}
	form.Set("client_id", client.clientID)
	var response TokenResponse
	err = client.do(ctx, "exchangeOAuthToken", "POST", "/oauth/token", form, requestAuth{}, &response, false, options...)
	if err != nil {
		return nil, sanitizeConfidentialError(err, sensitive...)
	}
	if response.AccessToken == "" || !strings.EqualFold(response.TokenType, "bearer") || response.ExpiresIn <= 0 {
		return nil, &aether.Error{Code: "invalid_token_response", Message: "Identity returned an invalid OAuth token response"}
	}
	return &response, nil
}

// RevokeOAuthToken revokes the presented token for this registered public client.
func (client *Client) RevokeOAuthToken(ctx context.Context, request TokenHandleRequest, options ...aether.RequestOption) error {
	if client.clientID == "" {
		return fmt.Errorf("identity: public client ID is required")
	}
	form, sensitive, err := tokenHandleForm(request)
	if err != nil {
		return err
	}
	form.Set("client_id", client.clientID)
	return sanitizeConfidentialError(client.do(ctx, "revokeOAuthToken", "POST", "/oauth/revoke", form, requestAuth{}, nil, false, options...), sensitive...)
}

func tokenRequestForm(request TokenRequest) (url.Values, []string, error) {
	form := make(url.Values)
	form.Set("grant_type", string(request.GrantType))
	sensitive := []string{request.Code, request.CodeVerifier, request.RefreshToken}
	switch request.GrantType {
	case GrantAuthorizationCode:
		if request.Code == "" || request.RedirectURI == "" || request.CodeVerifier == "" {
			return nil, nil, fmt.Errorf("identity: authorization code, redirect URI, and code verifier are required")
		}
		if !codeVerifierPattern.MatchString(request.CodeVerifier) {
			return nil, nil, fmt.Errorf("identity: code verifier must be 43 to 128 base64url characters")
		}
		form.Set("code", request.Code)
		form.Set("redirect_uri", request.RedirectURI)
		form.Set("code_verifier", request.CodeVerifier)
	case GrantRefreshToken:
		if request.RefreshToken == "" {
			return nil, nil, fmt.Errorf("identity: refresh token is required")
		}
		form.Set("refresh_token", request.RefreshToken)
	case GrantClientCredentials:
		if request.Audience == "" || request.Scope == "" {
			return nil, nil, fmt.Errorf("identity: audience and scope are required")
		}
		form.Set("audience", request.Audience)
		form.Set("scope", request.Scope)
	default:
		return nil, nil, fmt.Errorf("identity: unsupported OAuth grant type")
	}
	return form, sensitive, nil
}

func tokenHandleForm(request TokenHandleRequest) (url.Values, []string, error) {
	if request.Token == "" {
		return nil, nil, fmt.Errorf("identity: token is required")
	}
	form := url.Values{"token": {request.Token}}
	if request.TokenTypeHint != "" {
		form.Set("token_type_hint", request.TokenTypeHint)
	}
	return form, []string{request.Token}, nil
}

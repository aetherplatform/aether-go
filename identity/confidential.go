//go:build !js

package identity

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aetherplatform/aether-go"
)

type ConfidentialConfig struct {
	Config
	ClientID     string
	ClientSecret string
}

type ConfidentialClient struct {
	client       *Client
	clientID     string
	clientSecret string
}

func NewConfidentialClient(config ConfidentialConfig) (*ConfidentialClient, error) {
	if strings.TrimSpace(config.ClientID) == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("identity: confidential client ID and secret are required")
	}
	client, err := NewClient(config.Config)
	if err != nil {
		return nil, err
	}
	return &ConfidentialClient{client: client, clientID: config.ClientID, clientSecret: config.ClientSecret}, nil
}

func (client *ConfidentialClient) ExchangeOAuthToken(ctx context.Context, request TokenRequest, options ...aether.RequestOption) (*TokenResponse, error) {
	form, sensitive, err := tokenRequestForm(request)
	if err != nil {
		return nil, err
	}
	var response TokenResponse
	err = client.client.do(ctx, "exchangeOAuthToken", "POST", "/oauth/token", form, client.auth(), &response, false, options...)
	if err != nil {
		return nil, sanitizeConfidentialError(err, append(sensitive, client.clientSecret)...)
	}
	if response.AccessToken == "" || !strings.EqualFold(response.TokenType, "bearer") || response.ExpiresIn <= 0 {
		return nil, &aether.Error{Code: "invalid_token_response", Message: "Identity returned an invalid OAuth token response"}
	}
	return &response, nil
}

func (client *ConfidentialClient) RevokeOAuthToken(ctx context.Context, request TokenHandleRequest, options ...aether.RequestOption) error {
	form, sensitive, err := tokenHandleForm(request)
	if err != nil {
		return err
	}
	err = client.client.do(ctx, "revokeOAuthToken", "POST", "/oauth/revoke", form, client.auth(), nil, false, options...)
	return sanitizeConfidentialError(err, append(sensitive, client.clientSecret)...)
}

func (client *ConfidentialClient) IntrospectOAuthToken(ctx context.Context, request TokenHandleRequest, options ...aether.RequestOption) (*IntrospectionResponse, error) {
	form, sensitive, err := tokenHandleForm(request)
	if err != nil {
		return nil, err
	}
	var response IntrospectionResponse
	err = client.client.do(ctx, "introspectOAuthToken", "POST", "/oauth/introspect", form, client.auth(), &response, false, options...)
	if err != nil {
		return nil, sanitizeConfidentialError(err, append(sensitive, client.clientSecret)...)
	}
	return &response, nil
}

func (client *ConfidentialClient) auth() requestAuth {
	return requestAuth{basicID: client.clientID, basicSecret: client.clientSecret}
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

func sanitizeConfidentialError(err error, secrets ...string) error {
	if err == nil {
		return nil
	}
	requestErr, ok := err.(*aether.Error)
	if !ok {
		return err
	}
	for _, secret := range secrets {
		if secret != "" {
			requestErr.Message = strings.ReplaceAll(requestErr.Message, secret, "[REDACTED]")
		}
	}
	requestErr.Details = nil
	return requestErr
}

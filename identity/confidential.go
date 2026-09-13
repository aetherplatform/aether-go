//go:build !js

package identity

import (
	"context"
	"fmt"
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

func (client *ConfidentialClient) StartPasswordless(ctx context.Context, request PasswordlessStartRequest, options ...aether.RequestOption) (*PasswordlessStartResponse, error) {
	return client.client.startPasswordless(ctx, request, client.auth(), options...)
}

func (client *ConfidentialClient) VerifyPasswordless(ctx context.Context, request PasswordlessVerifyRequest, options ...aether.RequestOption) (PasswordlessVerifyResponse, error) {
	return client.client.verifyPasswordless(ctx, request, client.auth(), options...)
}

func (client *ConfidentialClient) CompletePasswordless(ctx context.Context, request PasswordlessCompleteRequest, options ...aether.RequestOption) (PasswordlessCompleteResponse, error) {
	return client.client.completePasswordless(ctx, request, client.auth(), options...)
}

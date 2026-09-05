package identity

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

const (
	defaultTimeout    = 15 * time.Second
	defaultMaxRetries = 2
)

var codeChallengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43,128}$`)
var codeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
	MaxRetries int
	UserAgent  string
	OnRetry    func(aether.RetryEvent)
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	timeout    time.Duration
	maxRetries int
	userAgent  string
	onRetry    func(aether.RetryEvent)
}

func NewClient(config Config) (*Client, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" || (baseURL.Scheme != "http" && baseURL.Scheme != "https") {
		return nil, fmt.Errorf("identity: base URL must be an absolute HTTP URL")
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	if timeout < 0 {
		return nil, fmt.Errorf("identity: timeout must not be negative")
	}
	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}
	if maxRetries < 0 {
		return nil, fmt.Errorf("identity: maximum retries must not be negative")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	userAgent := strings.TrimSpace(config.UserAgent)
	if userAgent == "" {
		userAgent = "aether-go/" + aether.Version
	}
	return &Client{baseURL: baseURL, httpClient: httpClient, timeout: timeout, maxRetries: maxRetries, userAgent: userAgent, onRetry: config.OnRetry}, nil
}

func (client *Client) AuthorizationURL(request AuthorizationRequest) (string, error) {
	if strings.TrimSpace(request.ClientID) == "" || strings.TrimSpace(request.Scope) == "" || strings.TrimSpace(request.State) == "" {
		return "", fmt.Errorf("identity: client ID, scope, and state are required")
	}
	redirectURI, err := url.Parse(request.RedirectURI)
	if err != nil || redirectURI.Scheme == "" {
		return "", fmt.Errorf("identity: redirect URI must be absolute")
	}
	if !codeChallengePattern.MatchString(request.CodeChallenge) {
		return "", fmt.Errorf("identity: code challenge must be 43 to 128 base64url characters")
	}
	target := client.baseURL.ResolveReference(&url.URL{Path: "/oauth/authorize"})
	query := target.Query()
	query.Set("response_type", "code")
	query.Set("client_id", request.ClientID)
	query.Set("redirect_uri", request.RedirectURI)
	query.Set("scope", request.Scope)
	query.Set("state", request.State)
	query.Set("code_challenge", request.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	if request.Nonce != "" {
		query.Set("nonce", request.Nonce)
	}
	target.RawQuery = query.Encode()
	return target.String(), nil
}

func (client *Client) GetOpenIDConfiguration(ctx context.Context, options ...aether.RequestOption) (*OpenIDConfiguration, error) {
	var response OpenIDConfiguration
	err := client.do(ctx, "getOpenIdConfiguration", http.MethodGet, "/.well-known/openid-configuration", nil, requestAuth{}, &response, true, options...)
	if err == nil && (response.Issuer == "" || response.AuthorizationEndpoint == "" || response.TokenEndpoint == "" || response.JWKSURI == "" || len(response.ResponseTypesSupported) == 0 || len(response.GrantTypesSupported) == 0 || len(response.CodeChallengeMethodsSupported) == 0) {
		err = invalidIdentityResponse("OpenID configuration")
	}
	return &response, err
}

func (client *Client) GetOAuthJWKS(ctx context.Context, options ...aether.RequestOption) (*JWKS, error) {
	var response JWKS
	err := client.do(ctx, "getOAuthJwks", http.MethodGet, "/.well-known/jwks.json", nil, requestAuth{}, &response, true, options...)
	if err == nil && len(response.Keys) == 0 {
		err = invalidIdentityResponse("JWKS")
	}
	return &response, err
}

func (client *Client) ListOAuthScopes(ctx context.Context, options ...aether.RequestOption) (*ScopeCatalog, error) {
	var response ScopeCatalog
	err := client.do(ctx, "listOAuthScopes", http.MethodGet, "/.well-known/scopes", nil, requestAuth{}, &response, true, options...)
	return &response, err
}

func (client *Client) GetUserInfo(ctx context.Context, tokenProvider aether.TokenProvider, options ...aether.RequestOption) (UserInfo, error) {
	if tokenProvider == nil {
		return nil, fmt.Errorf("identity: user access token provider is required")
	}
	refreshed := false
	for {
		token, err := tokenProvider.Token(ctx)
		if err != nil {
			return nil, err
		}
		if token.Value == "" {
			return nil, &aether.Error{Code: "invalid_access_token", Message: "Token provider returned an empty user access token"}
		}
		var response UserInfo
		err = client.do(ctx, "getUserInfo", http.MethodGet, "/oauth/userinfo", nil, requestAuth{bearer: token.Value}, &response, true, options...)
		var requestErr *aether.Error
		if errors.As(err, &requestErr) && requestErr.StatusCode == http.StatusUnauthorized && !refreshed {
			if invalidator, ok := tokenProvider.(aether.TokenInvalidator); ok {
				invalidator.Invalidate()
				refreshed = true
				continue
			}
		}
		if err == nil && response.Subject() == "" {
			err = invalidIdentityResponse("userinfo")
		}
		return response, err
	}
}

type requestAuth struct {
	bearer      string
	basicID     string
	basicSecret string
}

func (client *Client) do(ctx context.Context, operation, method, path string, form url.Values, auth requestAuth, result any, retrySafe bool, requestOptions ...aether.RequestOption) error {
	if ctx == nil {
		return fmt.Errorf("identity: context is required")
	}
	metadata := applyRequestOptions(requestOptions)
	if metadata.RequestID == "" {
		metadata.RequestID = randomRequestID()
	}
	target := client.baseURL.ResolveReference(&url.URL{Path: path})
	encodedForm := ""
	if form != nil {
		encodedForm = form.Encode()
	}
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return &aether.Error{Code: "request_aborted", Message: "Identity request was aborted", Attempts: attempt, Cause: err}
		}
		requestContext, cancel := context.WithTimeout(ctx, client.timeout)
		var body io.Reader
		if form != nil {
			body = bytes.NewBufferString(encodedForm)
		}
		request, err := http.NewRequestWithContext(requestContext, method, target.String(), body)
		if err != nil {
			cancel()
			return &aether.Error{Code: "request_build_failed", Message: "Failed to build Identity request", Attempts: attempt, Cause: err}
		}
		request.Header.Set("accept", "application/json")
		request.Header.Set("user-agent", client.userAgent)
		if metadata.RequestID != "" {
			request.Header.Set("x-request-id", metadata.RequestID)
		}
		if metadata.CorrelationID != "" {
			request.Header.Set("x-correlation-id", metadata.CorrelationID)
		}
		if form != nil {
			request.Header.Set("content-type", "application/x-www-form-urlencoded")
		}
		if auth.bearer != "" {
			request.Header.Set("authorization", "Bearer "+auth.bearer)
		}
		if auth.basicID != "" {
			request.SetBasicAuth(auth.basicID, auth.basicSecret)
		}
		response, sendErr := client.httpClient.Do(request)
		requestTimedOut := errors.Is(requestContext.Err(), context.DeadlineExceeded)
		if sendErr != nil {
			cancel()
			requestErr := &aether.Error{Code: "network_error", Message: "Identity request failed", Attempts: attempt, Cause: sendErr}
			if errors.Is(ctx.Err(), context.Canceled) {
				requestErr.Code, requestErr.Message, requestErr.Cause = "request_aborted", "Identity request was aborted", ctx.Err()
			} else if requestTimedOut {
				requestErr.Code, requestErr.Message = "request_timeout", "Identity request timed out"
			}
			if !client.retry(ctx, operation, attempt, requestErr, retrySafe) {
				if ctx.Err() != nil {
					return &aether.Error{Code: "request_aborted", Message: "Identity request was aborted", Attempts: attempt, Cause: ctx.Err()}
				}
				return requestErr
			}
			continue
		}
		responseBody, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		requestContextErr := requestContext.Err()
		cancel()
		if readErr != nil {
			if ctx.Err() != nil {
				return &aether.Error{StatusCode: response.StatusCode, Code: "request_aborted", Message: "Identity request was aborted", Attempts: attempt, Cause: ctx.Err()}
			}
			if errors.Is(requestContextErr, context.DeadlineExceeded) {
				return &aether.Error{StatusCode: response.StatusCode, Code: "request_timeout", Message: "Identity request timed out", Attempts: attempt, Cause: readErr}
			}
			return &aether.Error{StatusCode: response.StatusCode, Code: "response_read_failed", Message: "Failed to read Identity response", Attempts: attempt, Cause: readErr}
		}
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			if result == nil || len(responseBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(responseBody, result); err != nil {
				return &aether.Error{StatusCode: response.StatusCode, Code: "invalid_response", Message: "Identity returned an invalid JSON response", Attempts: attempt, Cause: err}
			}
			return nil
		}
		requestErr := transport.ErrorFromResponse(response, responseBody, attempt)
		if !client.retry(ctx, operation, attempt, requestErr, retrySafe) {
			if ctx.Err() != nil {
				return &aether.Error{Code: "request_aborted", Message: "Identity request was aborted", Attempts: attempt, Cause: ctx.Err()}
			}
			return requestErr
		}
	}
}

func (client *Client) retry(ctx context.Context, operation string, attempt int, requestErr *aether.Error, retrySafe bool) bool {
	if !retrySafe || attempt > client.maxRetries || !transport.Retryable(requestErr) || ctx.Err() != nil {
		return false
	}
	delay := transport.RetryDelay(requestErr, attempt, time.Now())
	if client.onRetry != nil {
		client.onRetry(aether.RetryEvent{Operation: operation, Attempt: attempt, MaxAttempts: client.maxRetries + 1, Delay: delay, StatusCode: requestErr.StatusCode, Code: requestErr.Code, RequestID: requestErr.RequestID})
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

type requestMetadata struct {
	RequestID     string
	CorrelationID string
}

func applyRequestOptions(options []aether.RequestOption) requestMetadata {
	var values transport.RequestOptions
	for _, option := range options {
		if option != nil {
			option(&values)
		}
	}
	return requestMetadata{RequestID: values.RequestID, CorrelationID: values.CorrelationID}
}

func randomRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(value[:])
}

func invalidIdentityResponse(name string) *aether.Error {
	return &aether.Error{Code: "invalid_response", Message: "Identity returned an invalid " + name + " response"}
}

//go:build !js

package clientcredentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

const (
	defaultTimeout    = 10 * time.Second
	defaultExpirySkew = 30 * time.Second
	defaultMaxRetries = 1
)

type Config struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Audience     string
	Capabilities []string
	HTTPClient   *http.Client
	Timeout      time.Duration
	// MaxRetries defaults to 1 when zero. Negative values are invalid.
	MaxRetries int
	// DisableRetries takes precedence over a non-negative MaxRetries value.
	DisableRetries bool
	ExpirySkew     time.Duration
	OnRetry        func(aether.RetryEvent)
}

type Provider struct {
	config     Config
	tokenURL   *url.URL
	httpClient *http.Client
	mu         sync.Mutex
	cached     aether.AccessToken
	inFlight   *tokenCall
}

type tokenCall struct {
	done    chan struct{}
	cancel  context.CancelFunc
	waiters int
	token   aether.AccessToken
	err     error
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func New(config Config) (*Provider, error) {
	tokenURL, err := url.Parse(config.TokenURL)
	if err != nil || tokenURL.Scheme == "" || tokenURL.Host == "" {
		return nil, fmt.Errorf("aether: token URL must be an absolute HTTP URL")
	}
	if tokenURL.Scheme != "http" && tokenURL.Scheme != "https" {
		return nil, fmt.Errorf("aether: token URL must use HTTP or HTTPS")
	}
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, fmt.Errorf("aether: client ID is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("aether: client secret is required")
	}
	if strings.TrimSpace(config.Audience) == "" {
		return nil, fmt.Errorf("aether: audience is required")
	}
	if len(config.Capabilities) == 0 {
		return nil, fmt.Errorf("aether: at least one capability is required")
	}
	for _, capability := range config.Capabilities {
		if strings.TrimSpace(capability) == "" {
			return nil, fmt.Errorf("aether: capabilities must not contain empty values")
		}
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("aether: timeout must not be negative")
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetries
	}
	if config.MaxRetries < 0 {
		return nil, fmt.Errorf("aether: maximum retries must not be negative")
	}
	if config.DisableRetries {
		config.MaxRetries = 0
	}
	if config.ExpirySkew == 0 {
		config.ExpirySkew = defaultExpirySkew
	}
	if config.ExpirySkew < 0 {
		return nil, fmt.Errorf("aether: expiry skew must not be negative")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Provider{config: config, tokenURL: tokenURL, httpClient: httpClient}, nil
}

func (provider *Provider) Token(ctx context.Context) (aether.AccessToken, error) {
	if ctx == nil {
		return aether.AccessToken{}, fmt.Errorf("aether: context is required")
	}
	provider.mu.Lock()
	if provider.cached.Value != "" && provider.cached.ExpiresAt.Add(-provider.config.ExpirySkew).After(time.Now()) {
		token := provider.cached
		provider.mu.Unlock()
		return token, nil
	}
	if provider.inFlight == nil {
		acquisitionContext, cancel := context.WithTimeout(
			context.Background(),
			provider.config.Timeout*time.Duration(provider.config.MaxRetries+1)+time.Second*time.Duration(provider.config.MaxRetries),
		)
		provider.inFlight = &tokenCall{done: make(chan struct{}), cancel: cancel}
		call := provider.inFlight
		go provider.acquire(acquisitionContext, call)
	}
	call := provider.inFlight
	call.waiters++
	provider.mu.Unlock()

	select {
	case <-ctx.Done():
		provider.mu.Lock()
		if provider.inFlight == call {
			call.waiters--
			if call.waiters == 0 {
				call.cancel()
			}
		}
		provider.mu.Unlock()
		return aether.AccessToken{}, &aether.Error{Code: "request_aborted", Message: "Token request was aborted", Cause: ctx.Err()}
	case <-call.done:
		return call.token, call.err
	}
}

func (provider *Provider) Invalidate() {
	provider.mu.Lock()
	provider.cached = aether.AccessToken{}
	provider.mu.Unlock()
}

func (provider *Provider) acquire(ctx context.Context, call *tokenCall) {
	defer call.cancel()
	call.token, call.err = provider.issue(ctx)

	provider.mu.Lock()
	if call.err == nil {
		provider.cached = call.token
	}
	if provider.inFlight == call {
		provider.inFlight = nil
	}
	close(call.done)
	provider.mu.Unlock()
}

func (provider *Provider) issue(ctx context.Context) (aether.AccessToken, error) {
	for attempt := 1; ; attempt++ {
		token, requestErr := provider.issueOnce(ctx, attempt)
		if requestErr == nil {
			return token, nil
		}
		if attempt > provider.config.MaxRetries || !transport.Retryable(requestErr) {
			requestErr.Attempts = attempt
			return aether.AccessToken{}, requestErr
		}
		delay := transport.RetryDelay(requestErr, attempt, time.Now())
		if provider.config.OnRetry != nil {
			provider.config.OnRetry(aether.RetryEvent{
				Operation:   "token_acquisition",
				Attempt:     attempt,
				MaxAttempts: provider.config.MaxRetries + 1,
				Delay:       delay,
				StatusCode:  requestErr.StatusCode,
				Code:        requestErr.Code,
				RequestID:   requestErr.RequestID,
			})
		}
		if err := wait(ctx, delay); err != nil {
			return aether.AccessToken{}, &aether.Error{Code: "request_aborted", Message: "Token request was aborted", Attempts: attempt, Cause: err}
		}
	}
}

func (provider *Provider) issueOnce(parent context.Context, attempt int) (aether.AccessToken, *aether.Error) {
	ctx, cancel := context.WithTimeout(parent, provider.config.Timeout)
	defer cancel()
	body := url.Values{
		"grant_type": {"client_credentials"},
		"audience":   {provider.config.Audience},
		"scope":      {strings.Join(provider.config.Capabilities, " ")},
	}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.tokenURL.String(), strings.NewReader(body))
	if err != nil {
		return aether.AccessToken{}, &aether.Error{Code: "token_request_failed", Message: "Token request failed", Attempts: attempt, Cause: err}
	}
	request.SetBasicAuth(provider.config.ClientID, provider.config.ClientSecret)
	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/x-www-form-urlencoded")

	response, err := provider.httpClient.Do(request)
	if err != nil {
		if errors.Is(parent.Err(), context.Canceled) {
			return aether.AccessToken{}, &aether.Error{Code: "request_aborted", Message: "Token request was aborted", Attempts: attempt, Cause: parent.Err()}
		}
		code := "token_request_failed"
		message := "Token request failed"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code = "request_timeout"
			message = "Token request timed out"
		}
		return aether.AccessToken{}, &aether.Error{Code: code, Message: message, Attempts: attempt, Cause: err}
	}
	responseBody, readErr := io.ReadAll(response.Body)
	response.Body.Close()
	if readErr != nil {
		return aether.AccessToken{}, &aether.Error{StatusCode: response.StatusCode, Code: "response_read_failed", Message: "Failed to read token response", Attempts: attempt, Cause: readErr}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseErr := transport.ErrorFromResponse(response, responseBody, attempt)
		responseErr.Message = strings.ReplaceAll(responseErr.Message, provider.config.ClientSecret, "[REDACTED]")
		return aether.AccessToken{}, responseErr
	}
	var payload tokenResponse
	if json.Unmarshal(responseBody, &payload) != nil || payload.AccessToken == "" || !strings.EqualFold(payload.TokenType, "bearer") || payload.ExpiresIn <= 0 {
		return aether.AccessToken{}, &aether.Error{StatusCode: response.StatusCode, Code: "invalid_token_response", Message: "Identity returned an invalid client credentials response", Attempts: attempt}
	}
	return aether.AccessToken{Value: payload.AccessToken, ExpiresAt: time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)}, nil
}

func wait(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package transport

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
	"strings"
	"time"
)

const (
	defaultTimeout    = 15 * time.Second
	defaultMaxRetries = 2
)

type Client struct {
	baseURL       *url.URL
	tokenProvider TokenProvider
	httpClient    *http.Client
	timeout       time.Duration
	maxRetries    int
	userAgent     string
	onRetry       func(RetryEvent)
}

func NewClient(config Config, defaultAgent string) (*Client, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("aether: base URL must be an absolute HTTP URL")
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("aether: base URL must use HTTP or HTTPS")
	}
	if config.TokenProvider == nil {
		return nil, fmt.Errorf("aether: token provider is required")
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	if timeout < 0 {
		return nil, fmt.Errorf("aether: timeout must not be negative")
	}
	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}
	if maxRetries < 0 {
		return nil, fmt.Errorf("aether: maximum retries must not be negative")
	}
	userAgent := strings.TrimSpace(config.UserAgent)
	if userAgent == "" {
		userAgent = defaultAgent
	}
	return &Client{
		baseURL:       baseURL,
		tokenProvider: config.TokenProvider,
		httpClient:    clientOrDefault(config.HTTPClient),
		timeout:       timeout,
		maxRetries:    maxRetries,
		userAgent:     userAgent,
		onRetry:       config.OnRetry,
	}, nil
}

func (client *Client) Execute(
	ctx context.Context,
	operation Operation,
	pathValues map[string]string,
	query url.Values,
	body any,
	result any,
	requestOptions ...RequestOption,
) error {
	if ctx == nil {
		return fmt.Errorf("aether: context is required")
	}
	options := RequestOptions{}
	for _, apply := range requestOptions {
		if apply != nil {
			apply(&options)
		}
	}
	requestID := strings.TrimSpace(options.RequestID)
	if requestID == "" {
		requestID = randomRequestID()
	}
	requestBody, err := marshalBody(body)
	if err != nil {
		return fmt.Errorf("aether: encode %s request: %w", operation.Name, err)
	}
	requestURL, err := buildURL(client.baseURL, operation.Path, pathValues, query)
	if err != nil {
		return err
	}

	retrySafe := operationRetrySafe(operation, options, body)
	refreshed := false
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return abortedError(err, attempt)
		}
		token, tokenErr := client.tokenProvider.Token(ctx)
		if tokenErr != nil {
			return tokenErr
		}
		if strings.TrimSpace(token.Value) == "" {
			return &Error{Code: "invalid_access_token", Message: "Token provider returned an empty access token", Attempts: attempt}
		}

		response, requestContext, cancelRequest, sendErr := client.send(ctx, operation, requestURL, requestBody, token.Value, requestID, options, attempt)
		if sendErr != nil {
			if !retrySafe || attempt > client.maxRetries || !Retryable(sendErr) {
				return sendErr
			}
			client.waitForRetry(ctx, operation.Name, attempt, sendErr, requestID)
			if err := wait(ctx, RetryDelay(sendErr, attempt, time.Now())); err != nil {
				return abortedError(err, attempt)
			}
			continue
		}

		responseBody, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		requestContextErr := requestContext.Err()
		cancelRequest()
		if readErr != nil {
			if ctx.Err() != nil {
				return abortedError(ctx.Err(), attempt)
			}
			if errors.Is(requestContextErr, context.DeadlineExceeded) {
				return &Error{StatusCode: response.StatusCode, Code: "request_timeout", Message: "Aether request timed out", RequestID: response.Header.Get("x-request-id"), Attempts: attempt, Cause: readErr}
			}
			return &Error{StatusCode: response.StatusCode, Code: "response_read_failed", Message: "Failed to read Aether response", RequestID: response.Header.Get("x-request-id"), Attempts: attempt, Cause: readErr}
		}
		if successStatus(operation.SuccessStatuses, response.StatusCode) {
			if result == nil || response.StatusCode == http.StatusNoContent || len(responseBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(responseBody, result); err != nil {
				return &Error{StatusCode: response.StatusCode, Code: "invalid_response", Message: "Aether returned an invalid JSON response", RequestID: response.Header.Get("x-request-id"), Attempts: attempt, Cause: err}
			}
			return nil
		}

		responseErr := ErrorFromResponse(response, responseBody, attempt)
		if response.StatusCode == http.StatusUnauthorized && !refreshed && retrySafe {
			if invalidator, ok := client.tokenProvider.(TokenInvalidator); ok {
				invalidator.Invalidate()
				refreshed = true
				if attempt <= client.maxRetries {
					client.emitRetry(operation.Name, attempt, 0, responseErr, requestID)
					continue
				}
			}
		}
		if !retrySafe || attempt > client.maxRetries || !Retryable(responseErr) {
			return responseErr
		}
		delay := RetryDelay(responseErr, attempt, time.Now())
		client.emitRetry(operation.Name, attempt, delay, responseErr, requestID)
		if err := wait(ctx, delay); err != nil {
			return abortedError(err, attempt)
		}
	}
}

func (client *Client) send(
	ctx context.Context,
	operation Operation,
	requestURL string,
	body []byte,
	token string,
	requestID string,
	options RequestOptions,
	attempt int,
) (*http.Response, context.Context, context.CancelFunc, *Error) {
	requestContext := ctx
	cancel := func() {}
	if client.timeout > 0 {
		requestContext, cancel = context.WithTimeout(ctx, client.timeout)
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(requestContext, operation.Method, requestURL, reader)
	if err != nil {
		cancel()
		return nil, requestContext, func() {}, &Error{Code: "request_build_failed", Message: "Failed to build Aether request", RequestID: requestID, Attempts: attempt, Cause: err}
	}
	request.Header.Set("accept", "application/json")
	request.Header.Set("authorization", "Bearer "+token)
	request.Header.Set("user-agent", client.userAgent)
	request.Header.Set("x-request-id", requestID)
	if options.CorrelationID != "" {
		request.Header.Set("x-correlation-id", options.CorrelationID)
	}
	if options.IdempotencyKey != "" {
		request.Header.Set("idempotency-key", options.IdempotencyKey)
	}
	if body != nil {
		request.Header.Set("content-type", "application/json")
	}

	response, err := client.httpClient.Do(request)
	if err == nil {
		return response, requestContext, cancel, nil
	}
	cancel()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil, requestContext, func() {}, abortedError(ctx.Err(), attempt)
	}
	if errors.Is(requestContext.Err(), context.DeadlineExceeded) {
		return nil, requestContext, func() {}, &Error{Code: "request_timeout", Message: "Aether request timed out", RequestID: requestID, Attempts: attempt, Cause: err}
	}
	return nil, requestContext, func() {}, &Error{Code: "network_error", Message: "Aether request failed", RequestID: requestID, Attempts: attempt, Cause: err}
}

func (client *Client) waitForRetry(ctx context.Context, operation string, attempt int, requestErr *Error, requestID string) {
	delay := RetryDelay(requestErr, attempt, time.Now())
	client.emitRetry(operation, attempt, delay, requestErr, requestID)
}

func (client *Client) emitRetry(operation string, attempt int, delay time.Duration, requestErr *Error, requestID string) {
	if client.onRetry == nil {
		return
	}
	client.onRetry(RetryEvent{
		Operation:   operation,
		Attempt:     attempt,
		MaxAttempts: client.maxRetries + 1,
		Delay:       delay,
		StatusCode:  requestErr.StatusCode,
		Code:        requestErr.Code,
		RequestID:   requestID,
	})
}

func operationRetrySafe(operation Operation, options RequestOptions, body any) bool {
	switch operation.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	switch operation.Idempotency {
	case IdempotencyRequired, IdempotencyOptional:
		return options.IdempotencyKey != ""
	case IdempotencyRequestField:
		return operation.RequestRetrySafe != nil && operation.RequestRetrySafe(body)
	default:
		return false
	}
}

func Retryable(err *Error) bool {
	if err == nil {
		return false
	}
	if err.StatusCode == 0 {
		return err.Code == "network_error" || err.Code == "request_timeout" || err.Code == "token_request_failed"
	}
	if err.StatusCode == http.StatusBadGateway || err.StatusCode == http.StatusServiceUnavailable || err.StatusCode == http.StatusGatewayTimeout {
		return true
	}
	return err.StatusCode == http.StatusTooManyRequests && err.Code == "rate_limited"
}

func buildURL(baseURL *url.URL, pathTemplate string, pathValues map[string]string, query url.Values) (string, error) {
	path := pathTemplate
	for {
		start := strings.IndexByte(path, '{')
		if start < 0 {
			break
		}
		endOffset := strings.IndexByte(path[start:], '}')
		if endOffset < 0 {
			return "", fmt.Errorf("aether: malformed path template")
		}
		end := start + endOffset
		name := path[start+1 : end]
		value, ok := pathValues[name]
		if !ok || value == "" {
			return "", fmt.Errorf("aether: missing path parameter %s", name)
		}
		path = path[:start] + url.PathEscape(value) + path[end+1:]
	}
	resolved, err := url.Parse(strings.TrimRight(baseURL.String(), "/") + path)
	if err != nil {
		return "", fmt.Errorf("aether: build request URL: %w", err)
	}
	resolved.RawQuery = query.Encode()
	return resolved.String(), nil
}

func marshalBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return json.Marshal(body)
}

func successStatus(statuses []int, status int) bool {
	for _, candidate := range statuses {
		if candidate == status {
			return true
		}
	}
	return false
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

func abortedError(cause error, attempts int) *Error {
	return &Error{Code: "request_aborted", Message: "Aether request was aborted", Attempts: attempts, Cause: cause}
}

func randomRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(value[:])
}

func clientOrDefault(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}

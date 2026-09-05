package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type testTokenProvider struct {
	mu          sync.Mutex
	tokens      []string
	index       int
	invalidated int
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func (provider *testTokenProvider) Token(context.Context) (AccessToken, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	index := provider.index
	if index >= len(provider.tokens) {
		index = len(provider.tokens) - 1
	}
	return AccessToken{Value: provider.tokens[index], ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (provider *testTokenProvider) Invalidate() {
	provider.mu.Lock()
	provider.invalidated++
	if provider.index < len(provider.tokens)-1 {
		provider.index++
	}
	provider.mu.Unlock()
}

func TestClientRefreshesOnceForSafeOperation(t *testing.T) {
	t.Parallel()
	provider := &testTokenProvider{tokens: []string{"stale", "fresh"}}
	requests := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Authorization") == "Bearer stale" {
			return response(http.StatusUnauthorized, `{}`), nil
		}
		return response(http.StatusOK, `{"ok":true}`), nil
	})}

	client, err := NewClient(Config{BaseURL: "https://api.useather.test", TokenProvider: provider, HTTPClient: httpClient}, "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		OK bool `json:"ok"`
	}
	err = client.Execute(context.Background(), Operation{Name: "read", Method: http.MethodGet, Path: "/resource", Idempotency: IdempotencyNotApplicable, SuccessStatuses: []int{200}}, nil, nil, nil, &result)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || requests != 2 || provider.invalidated != 1 {
		t.Fatalf("unexpected refresh result: ok=%v requests=%d invalidations=%d", result.OK, requests, provider.invalidated)
	}
}

func TestClientDoesNotRefreshUnsafeMutation(t *testing.T) {
	t.Parallel()
	provider := &testTokenProvider{tokens: []string{"stale", "fresh"}}
	requests := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		requests++
		return response(http.StatusUnauthorized, `{}`), nil
	})}

	client, err := NewClient(Config{BaseURL: "https://api.useather.test", TokenProvider: provider, HTTPClient: httpClient}, "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Execute(context.Background(), Operation{Name: "write", Method: http.MethodPost, Path: "/resource", Idempotency: IdempotencyRequired, SuccessStatuses: []int{200}}, nil, nil, map[string]string{"value": "one"}, nil)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if requests != 1 || provider.invalidated != 0 {
		t.Fatalf("unsafe mutation retried: requests=%d invalidations=%d", requests, provider.invalidated)
	}
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRetryVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Cases []struct {
			Name              string `json:"name"`
			Method            string `json:"method"`
			Idempotency       string `json:"idempotency"`
			HasIdempotencyKey bool   `json:"has_idempotency_key"`
			Status            int    `json:"status"`
			Code              string `json:"code"`
			Retry             bool   `json:"retry"`
		} `json:"cases"`
	}
	loadVector(t, "retry-vectors.json", &document)
	for _, testCase := range document.Cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			options := RequestOptions{}
			if testCase.HasIdempotencyKey {
				options.IdempotencyKey = "idem"
			}
			operation := Operation{Method: testCase.Method, Idempotency: IdempotencyPolicy(testCase.Idempotency)}
			got := operationRetrySafe(operation, options, nil) && Retryable(&Error{StatusCode: testCase.Status, Code: testCase.Code})
			if got != testCase.Retry {
				t.Fatalf("retry=%v, want %v", got, testCase.Retry)
			}
		})
	}
}

func TestBuildURLVector(t *testing.T) {
	t.Parallel()
	var document struct {
		Cases []struct {
			Name       string            `json:"name"`
			BaseURL    string            `json:"base_url"`
			Path       string            `json:"path"`
			PathValues map[string]string `json:"path_values"`
			Query      map[string]any    `json:"query"`
			Expected   string            `json:"expected"`
			Error      string            `json:"error"`
		} `json:"cases"`
	}
	loadVector(t, "transport-vectors.json", &document)
	for _, testCase := range document.Cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			base, err := url.Parse(testCase.BaseURL)
			if err != nil {
				t.Fatal(err)
			}
			query := make(url.Values)
			for name, value := range testCase.Query {
				query.Set(name, formatVectorValue(value))
			}
			got, err := buildURL(base, testCase.Path, testCase.PathValues, query)
			if testCase.Error != "" {
				if err == nil {
					t.Fatalf("expected %s", testCase.Error)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != testCase.Expected {
				t.Fatalf("URL=%q, want %q", got, testCase.Expected)
			}
		})
	}
}

func TestErrorVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Cases []struct {
			Name     string            `json:"name"`
			Status   int               `json:"status"`
			Headers  map[string]string `json:"headers"`
			Body     json.RawMessage   `json:"body"`
			Expected struct {
				Code               string `json:"code"`
				Message            string `json:"message"`
				RequestID          string `json:"request_id"`
				RetryAfterSeconds  int    `json:"retry_after_seconds"`
				RateLimitLimit     int    `json:"rate_limit_limit"`
				RateLimitRemaining int    `json:"rate_limit_remaining"`
				RateLimitReset     int64  `json:"rate_limit_reset"`
				QuotaReset         string `json:"quota_reset"`
			} `json:"expected"`
		} `json:"cases"`
	}
	loadVector(t, "error-vectors.json", &document)
	for _, testCase := range document.Cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			header := make(http.Header)
			for name, value := range testCase.Headers {
				header.Set(name, value)
			}
			got := ErrorFromResponse(&http.Response{StatusCode: testCase.Status, Header: header}, testCase.Body, 1)
			if got.Code != testCase.Expected.Code || got.Message != testCase.Expected.Message || got.RequestID != testCase.Expected.RequestID {
				t.Fatalf("error=%+v", got)
			}
			if got.Retry.RetryAfterSeconds != testCase.Expected.RetryAfterSeconds || got.Retry.RateLimitLimit != testCase.Expected.RateLimitLimit || got.Retry.RateLimitRemaining != testCase.Expected.RateLimitRemaining || got.Retry.RateLimitReset != testCase.Expected.RateLimitReset || got.Retry.QuotaReset != testCase.Expected.QuotaReset {
				t.Fatalf("retry metadata=%+v", got.Retry)
			}
		})
	}
}

func formatVectorValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return fmt.Sprintf("%g", typed)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func loadVector(t *testing.T, name string, target any) {
	t.Helper()
	paths := []string{
		filepath.Join("..", "..", "..", "..", "contracts", "sdk", "v1", name),
		filepath.Join("..", "..", "contracts", "sdk", "v1", name),
	}
	var data []byte
	var err error
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("read vector %s: %v", name, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

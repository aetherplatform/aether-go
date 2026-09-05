package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type errorEnvelope struct {
	Error json.RawMessage `json:"error"`
}

type structuredError struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	Details   json.RawMessage `json:"details"`
}

func ErrorFromResponse(response *http.Response, body []byte, attempts int) *Error {
	code := "request_failed"
	message := fmt.Sprintf("Aether request failed with HTTP %d", response.StatusCode)
	requestID := response.Header.Get("x-request-id")
	var details json.RawMessage

	var envelope errorEnvelope
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Error) > 0 {
		var structured structuredError
		if json.Unmarshal(envelope.Error, &structured) == nil && structured.Code != "" {
			code = structured.Code
			if structured.Message != "" {
				message = structured.Message
			}
			if requestID == "" {
				requestID = structured.RequestID
			}
			details = structured.Details
		} else {
			var simple string
			if json.Unmarshal(envelope.Error, &simple) == nil && simple != "" {
				code = simple
			}
		}
	}

	return &Error{
		StatusCode: response.StatusCode,
		Code:       code,
		Message:    message,
		RequestID:  requestID,
		Details:    details,
		Retry: RetryMetadata{
			RetryAfterSeconds:  retryAfterSeconds(response.Header, time.Now()),
			RateLimitLimit:     headerInt(response.Header, "ratelimit-limit"),
			RateLimitRemaining: headerInt(response.Header, "ratelimit-remaining"),
			RateLimitReset:     headerInt64(response.Header, "ratelimit-reset"),
			QuotaReset:         response.Header.Get("quota-reset"),
		},
		Attempts: attempts,
	}
}

func RetryDelay(err *Error, attempt int, _ time.Time) time.Duration {
	if err != nil {
		if err.Retry.RetryAfterSeconds > 0 {
			return time.Duration(err.Retry.RetryAfterSeconds) * time.Second
		}
	}
	delay := 100 * time.Millisecond * time.Duration(1<<max(attempt-1, 0))
	if delay > time.Second {
		return time.Second
	}
	return delay
}

func RetryAfter(header http.Header, now time.Time) time.Duration {
	raw := header.Get("retry-after")
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(raw); err == nil && retryAt.After(now) {
		return retryAt.Sub(now)
	}
	return 0
}

func headerInt(header http.Header, name string) int {
	value, _ := strconv.Atoi(header.Get(name))
	return value
}

func headerInt64(header http.Header, name string) int64 {
	value, _ := strconv.ParseInt(header.Get(name), 10, 64)
	return value
}

func retryAfterSeconds(header http.Header, now time.Time) int {
	delay := RetryAfter(header, now)
	if delay <= 0 {
		return 0
	}
	seconds := int(delay.Round(time.Second) / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

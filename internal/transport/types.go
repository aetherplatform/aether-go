package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type TokenProvider interface {
	Token(context.Context) (AccessToken, error)
}

type TokenInvalidator interface {
	Invalidate()
}

type RetryMetadata struct {
	RetryAfterSeconds  int
	RateLimitLimit     int
	RateLimitRemaining int
	RateLimitReset     int64
	QuotaReset         string
}

type RetryEvent struct {
	Operation   string
	Attempt     int
	MaxAttempts int
	Delay       time.Duration
	StatusCode  int
	Code        string
	RequestID   string
}

type Config struct {
	BaseURL       string
	TokenProvider TokenProvider
	HTTPClient    *http.Client
	Timeout       time.Duration
	MaxRetries    int
	UserAgent     string
	OnRetry       func(RetryEvent)
}

type Error struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Details    json.RawMessage
	Retry      RetryMetadata
	Attempts   int
	Cause      error
}

func (err *Error) Error() string {
	if err == nil {
		return ""
	}
	return err.Message
}

func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type IdempotencyPolicy string

const (
	IdempotencyRequired      IdempotencyPolicy = "required"
	IdempotencyOptional      IdempotencyPolicy = "optional"
	IdempotencyRequestField  IdempotencyPolicy = "request_field"
	IdempotencyUnsupported   IdempotencyPolicy = "unsupported"
	IdempotencyNotApplicable IdempotencyPolicy = "not_applicable"
)

type Operation struct {
	Name             string
	Method           string
	Path             string
	Idempotency      IdempotencyPolicy
	SuccessStatuses  []int
	RequestRetrySafe func(any) bool
}

type RequestOption func(*RequestOptions)

type RequestOptions struct {
	RequestID      string
	CorrelationID  string
	IdempotencyKey string
}

func WithRequestID(value string) RequestOption {
	return func(options *RequestOptions) {
		options.RequestID = value
	}
}

func WithCorrelationID(value string) RequestOption {
	return func(options *RequestOptions) {
		options.CorrelationID = value
	}
}

func WithIdempotencyKey(value string) RequestOption {
	return func(options *RequestOptions) {
		options.IdempotencyKey = value
	}
}

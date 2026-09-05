package aether

import "github.com/aetherplatform/aether-go/internal/transport"

type AccessToken = transport.AccessToken
type TokenProvider = transport.TokenProvider
type TokenInvalidator = transport.TokenInvalidator
type Config = transport.Config
type Error = transport.Error
type RetryMetadata = transport.RetryMetadata
type RetryEvent = transport.RetryEvent
type RequestOption = transport.RequestOption

var WithRequestID = transport.WithRequestID
var WithCorrelationID = transport.WithCorrelationID
var WithIdempotencyKey = transport.WithIdempotencyKey

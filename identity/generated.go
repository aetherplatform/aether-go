// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.
// Source: contracts/openapi/v1/identity.yaml

package identity

import (
	"github.com/aetherplatform/aether-go/internal/transport"
	"net/http"
)

type PasswordlessAuthorized struct {
	Code   string  `json:"code"`
	State  *string `json:"state"`
	Status string  `json:"status"`
}

type PasswordlessCompleteRequest struct {
	ClientID     *string                          `json:"client_id,omitempty"`
	CodeVerifier string                           `json:"code_verifier"`
	Consent      *PasswordlessRegistrationConsent `json:"consent,omitempty"`
	Continuation string                           `json:"continuation"`
	Decision     Decision                         `json:"decision"`
	Identifier   *string                          `json:"identifier,omitempty"`
	Profile      *PasswordlessRegistrationProfile `json:"profile,omitempty"`
}
type PasswordlessCompleteResponse interface{ isPasswordlessCompleteResponse() }

func (*PasswordlessAuthorized) isPasswordlessCompleteResponse() {}
func (*PasswordlessDenied) isPasswordlessCompleteResponse()     {}

type PasswordlessCompletionRequirements struct {
	AetherPrivacyVersion *string  `json:"aether_privacy_version"`
	AetherTermsVersion   *string  `json:"aether_terms_version"`
	Consent              bool     `json:"consent"`
	Registration         bool     `json:"registration"`
	Scopes               []string `json:"scopes"`
}

type PasswordlessCustomCompletionRequired struct {
	Continuation string                             `json:"continuation"`
	ExpiresIn    int64                              `json:"expires_in"`
	Requirements PasswordlessCompletionRequirements `json:"requirements"`
	Status       string                             `json:"status"`
}

type PasswordlessDenied struct {
	Error  string  `json:"error"`
	State  *string `json:"state"`
	Status string  `json:"status"`
}

type PasswordlessError struct {
	Error struct {
		Code Code `json:"code"`
	} `json:"error"`
}

type PasswordlessHostedCompletionRequired struct {
	ContinuationURL string `json:"continuation_url"`
	ExpiresIn       int64  `json:"expires_in"`
	Status          string `json:"status"`
}

type PasswordlessRegistrationConsent struct {
	AetherPrivacyVersion string `json:"aether_privacy_version"`
	AetherTermsVersion   string `json:"aether_terms_version"`
}

type PasswordlessRegistrationProfile struct {
	DisplayName *string `json:"display_name,omitempty"`
}

type PasswordlessStartRequest struct {
	Channel             Channel `json:"channel"`
	ClientID            *string `json:"client_id,omitempty"`
	CodeChallenge       string  `json:"code_challenge"`
	CodeChallengeMethod string  `json:"code_challenge_method"`
	Identifier          string  `json:"identifier"`
	Nonce               *string `json:"nonce,omitempty"`
	RedirectURI         string  `json:"redirect_uri"`
	ResponseType        *string `json:"response_type,omitempty"`
	Scope               *string `json:"scope,omitempty"`
	State               *string `json:"state,omitempty"`
}

type PasswordlessStartResponse struct {
	ChallengeID          string `json:"challenge_id"`
	ExpiresIn            int64  `json:"expires_in"`
	ResendAfter          int64  `json:"resend_after"`
	Transaction          string `json:"transaction"`
	TransactionExpiresIn int64  `json:"transaction_expires_in"`
}

type PasswordlessVerifyRequest struct {
	ChallengeID  string  `json:"challenge_id"`
	Channel      Channel `json:"channel"`
	ClientID     *string `json:"client_id,omitempty"`
	Code         string  `json:"code"`
	CodeVerifier string  `json:"code_verifier"`
	Identifier   string  `json:"identifier"`
	Transaction  string  `json:"transaction"`
}
type PasswordlessVerifyResponse interface{ isPasswordlessVerifyResponse() }

func (*PasswordlessAuthorized) isPasswordlessVerifyResponse()               {}
func (*PasswordlessHostedCompletionRequired) isPasswordlessVerifyResponse() {}
func (*PasswordlessCustomCompletionRequired) isPasswordlessVerifyResponse() {}

type Alg string

const (
	AlgEdDSA Alg = "EdDSA"
	AlgRS256 Alg = "RS256"
)

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSms   Channel = "sms"
)

type Code string

const (
	CodeINVALIDREQUEST             Code = "INVALID_REQUEST"
	CodeINVALIDPASSWORDLESSREQUEST Code = "INVALID_PASSWORDLESS_REQUEST"
	CodeINVALIDCLIENT              Code = "INVALID_CLIENT"
	CodeORIGINNOTALLOWED           Code = "ORIGIN_NOT_ALLOWED"
	CodeCUSTOMCOMPLETIONNOTALLOWED Code = "CUSTOM_COMPLETION_NOT_ALLOWED"
	CodeRATELIMITED                Code = "RATE_LIMITED"
	CodeTEMPORARILYUNAVAILABLE     Code = "TEMPORARILY_UNAVAILABLE"
	CodeIDENTITYROLEUNAVAILABLE    Code = "IDENTITY_ROLE_UNAVAILABLE"
)

type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionDeny    Decision = "deny"
)

type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeRefreshToken      GrantType = "refresh_token"
	GrantTypeClientCredentials GrantType = "client_credentials"
)

type Kty string

const (
	KtyOKP Kty = "OKP"
	KtyRSA Kty = "RSA"
)

type TokenEndpointAuthMethodsSupportedItem string

const (
	TokenEndpointAuthMethodsSupportedItemClientSecretBasic TokenEndpointAuthMethodsSupportedItem = "client_secret_basic"
	TokenEndpointAuthMethodsSupportedItemClientSecretPost  TokenEndpointAuthMethodsSupportedItem = "client_secret_post"
	TokenEndpointAuthMethodsSupportedItemNone              TokenEndpointAuthMethodsSupportedItem = "none"
)

var (
	authorizeOperation              = transport.Operation{Name: "authorize", Method: http.MethodGet, Path: "/oauth/authorize", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int(nil)}
	completePasswordlessOperation   = transport.Operation{Name: "completePasswordless", Method: http.MethodPost, Path: "/v1/passwordless/complete", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	exchangeOAuthTokenOperation     = transport.Operation{Name: "exchangeOAuthToken", Method: http.MethodPost, Path: "/oauth/token", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	getOAuthJwksOperation           = transport.Operation{Name: "getOAuthJwks", Method: http.MethodGet, Path: "/.well-known/jwks.json", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getOpenIdConfigurationOperation = transport.Operation{Name: "getOpenIdConfiguration", Method: http.MethodGet, Path: "/.well-known/openid-configuration", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getUserInfoOperation            = transport.Operation{Name: "getUserInfo", Method: http.MethodGet, Path: "/oauth/userinfo", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	introspectOAuthTokenOperation   = transport.Operation{Name: "introspectOAuthToken", Method: http.MethodPost, Path: "/oauth/introspect", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	listOAuthScopesOperation        = transport.Operation{Name: "listOAuthScopes", Method: http.MethodGet, Path: "/.well-known/scopes", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	revokeOAuthTokenOperation       = transport.Operation{Name: "revokeOAuthToken", Method: http.MethodPost, Path: "/oauth/revoke", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	startPasswordlessOperation      = transport.Operation{Name: "startPasswordless", Method: http.MethodPost, Path: "/v1/passwordless/start", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{202}}
	verifyPasswordlessOperation     = transport.Operation{Name: "verifyPasswordless", Method: http.MethodPost, Path: "/v1/passwordless/verify", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
)
var allOperations = []transport.Operation{
	authorizeOperation,
	completePasswordlessOperation,
	exchangeOAuthTokenOperation,
	getOAuthJwksOperation,
	getOpenIdConfigurationOperation,
	getUserInfoOperation,
	introspectOAuthTokenOperation,
	listOAuthScopesOperation,
	revokeOAuthTokenOperation,
	startPasswordlessOperation,
	verifyPasswordlessOperation,
}

package identity

type TokenGrantType string

const (
	GrantAuthorizationCode TokenGrantType = "authorization_code"
	GrantRefreshToken      TokenGrantType = "refresh_token"
	GrantClientCredentials TokenGrantType = "client_credentials"
)

type AuthorizationRequest struct {
	ClientID      string
	RedirectURI   string
	Scope         string
	State         string
	Nonce         string
	CodeChallenge string
}

type TokenRequest struct {
	GrantType    TokenGrantType
	Code         string
	RedirectURI  string
	CodeVerifier string
	RefreshToken string
	Audience     string
	Scope        string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type TokenHandleRequest struct {
	Token         string
	TokenTypeHint string
}

type IntrospectionResponse struct {
	Active    bool    `json:"active"`
	ClientID  *string `json:"client_id,omitempty"`
	TokenType *string `json:"token_type,omitempty"`
	TokenUse  *string `json:"token_use,omitempty"`
	Scope     *string `json:"scope,omitempty"`
	Audience  *string `json:"aud,omitempty"`
	Issuer    *string `json:"iss,omitempty"`
	Subject   *string `json:"sub,omitempty"`
	ExpiresAt *int64  `json:"exp,omitempty"`
	IssuedAt  *int64  `json:"iat,omitempty"`
	JTI       *string `json:"jti,omitempty"`
}

type UserInfo map[string]any

func (info UserInfo) Subject() string {
	subject, _ := info["sub"].(string)
	return subject
}

type OpenIDConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint,omitempty"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
	JWKSURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	IDTokenSigningAlgorithmsSupported []string `json:"id_token_signing_alg_values_supported,omitempty"`
}

type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	Curve     string `json:"crv,omitempty"`
	X         string `json:"x,omitempty"`
	Modulus   string `json:"n,omitempty"`
	Exponent  string `json:"e,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type ScopeCatalog struct {
	Scopes []string `json:"scopes"`
}

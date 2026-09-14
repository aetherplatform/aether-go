package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

var otpPattern = regexp.MustCompile(`^[0-9]{6}$`)
var s256Pattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// StartPasswordless sends a code. Calling it again resends only after the user
// requests it and the returned ResendAfter interval has elapsed; it never retries.
func (client *Client) StartPasswordless(ctx context.Context, request PasswordlessStartRequest, options ...aether.RequestOption) (*PasswordlessStartResponse, error) {
	return client.startPasswordless(ctx, request, requestAuth{}, options...)
}

func (client *Client) VerifyPasswordless(ctx context.Context, request PasswordlessVerifyRequest, options ...aether.RequestOption) (PasswordlessVerifyResponse, error) {
	return client.verifyPasswordless(ctx, request, requestAuth{}, options...)
}

// CompletePasswordless is available only to clients approved for custom
// completion. Hosted-only policy is enforced by Aether.
func (client *Client) CompletePasswordless(ctx context.Context, request PasswordlessCompleteRequest, options ...aether.RequestOption) (PasswordlessCompleteResponse, error) {
	return client.completePasswordless(ctx, request, requestAuth{}, options...)
}

func (client *Client) startPasswordless(ctx context.Context, request PasswordlessStartRequest, auth requestAuth, options ...aether.RequestOption) (*PasswordlessStartResponse, error) {
	id, err := client.passwordlessClientID(request.ClientID, auth)
	if err != nil {
		return nil, err
	}
	request.ClientID, auth.publicID = &id, id
	if strings.TrimSpace(request.Identifier) == "" || !validChannel(request.Channel) || !s256Pattern.MatchString(request.CodeChallenge) {
		return nil, fmt.Errorf("identity: identifier, email/sms channel and S256 challenge are required")
	}
	if request.CodeChallengeMethod == "" {
		request.CodeChallengeMethod = "S256"
	}
	if request.CodeChallengeMethod != "S256" || (request.ResponseType != nil && *request.ResponseType != "code") {
		return nil, fmt.Errorf("identity: passwordless requires S256 and response type code")
	}
	if _, err := validateRedirectURI(request.RedirectURI); err != nil {
		return nil, err
	}
	if (request.State != nil && len(*request.State) > 1024) || (request.Nonce != nil && len(*request.Nonce) > 1024) {
		return nil, fmt.Errorf("identity: state and nonce must not exceed 1024 bytes")
	}
	var response PasswordlessStartResponse
	if err := client.passwordlessJSON(ctx, startPasswordlessOperation, request, auth, &response, options...); err != nil {
		return nil, err
	}
	if response.Transaction == "" || response.ChallengeID == "" || response.ExpiresIn != 300 || response.ResendAfter < 1 || response.ResendAfter > 300 || response.TransactionExpiresIn != 600 {
		return nil, invalidIdentityResponse("passwordless start")
	}
	return &response, nil
}

func (client *Client) verifyPasswordless(ctx context.Context, request PasswordlessVerifyRequest, auth requestAuth, options ...aether.RequestOption) (PasswordlessVerifyResponse, error) {
	id, err := client.passwordlessClientID(request.ClientID, auth)
	if err != nil {
		return nil, err
	}
	request.ClientID, auth.publicID = &id, id
	if request.Transaction == "" || request.ChallengeID == "" || strings.TrimSpace(request.Identifier) == "" || !validChannel(request.Channel) || !otpPattern.MatchString(request.Code) || !codeVerifierPattern.MatchString(request.CodeVerifier) {
		return nil, fmt.Errorf("identity: transaction, challenge, identifier, channel, six-digit code and PKCE verifier are required")
	}
	var raw json.RawMessage
	if err := client.passwordlessJSON(ctx, verifyPasswordlessOperation, request, auth, &raw, options...); err != nil {
		return nil, err
	}
	outcome, err := decodePasswordlessOutcome(raw)
	if err != nil {
		return nil, err
	}
	response, ok := outcome.(PasswordlessVerifyResponse)
	if !ok {
		return nil, invalidIdentityResponse("passwordless verification")
	}
	return response, nil
}

func (client *Client) completePasswordless(ctx context.Context, request PasswordlessCompleteRequest, auth requestAuth, options ...aether.RequestOption) (PasswordlessCompleteResponse, error) {
	id, err := client.passwordlessClientID(request.ClientID, auth)
	if err != nil {
		return nil, err
	}
	request.ClientID, auth.publicID = &id, id
	if request.Continuation == "" || !codeVerifierPattern.MatchString(request.CodeVerifier) || (request.Decision != DecisionApprove && request.Decision != DecisionDeny) {
		return nil, fmt.Errorf("identity: continuation, PKCE verifier and approve/deny decision are required")
	}
	var raw json.RawMessage
	if err := client.passwordlessJSON(ctx, completePasswordlessOperation, request, auth, &raw, options...); err != nil {
		return nil, err
	}
	outcome, err := decodePasswordlessOutcome(raw)
	if err != nil {
		return nil, err
	}
	response, ok := outcome.(PasswordlessCompleteResponse)
	if !ok {
		return nil, invalidIdentityResponse("passwordless completion")
	}
	return response, nil
}

func (client *Client) passwordlessClientID(explicit *string, auth requestAuth) (string, error) {
	id := client.clientID
	if auth.basicID != "" {
		id = auth.basicID
	}
	if explicit != nil {
		if id != "" && id != *explicit {
			return "", fmt.Errorf("identity: request client ID must match the configured client")
		}
		id = *explicit
	}
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("identity: registered client ID is required")
	}
	return id, nil
}

func validChannel(channel Channel) bool { return channel == ChannelEmail || channel == ChannelSms }

func (client *Client) passwordlessJSON(ctx context.Context, op transport.Operation, request any, auth requestAuth, result any, options ...aether.RequestOption) error {
	encoded, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("identity: invalid passwordless request")
	}
	return client.doBody(ctx, op.Name, http.MethodPost, op.Path, encoded, "application/json", auth, result, false, options...)
}

func decodePasswordlessOutcome(raw json.RawMessage) (any, error) {
	invalid := func() (any, error) { return nil, invalidIdentityResponse("passwordless outcome") }
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return invalid()
	}
	var status string
	if json.Unmarshal(fields["status"], &status) != nil {
		return invalid()
	}
	switch status {
	case "authorized":
		var response PasswordlessAuthorized
		if json.Unmarshal(raw, &response) != nil || response.Code == "" || fields["state"] == nil {
			return invalid()
		}
		return &response, nil
	case "hosted_completion_required":
		var response PasswordlessHostedCompletionRequired
		if json.Unmarshal(raw, &response) != nil || response.ExpiresIn <= 0 || response.ExpiresIn > 300 {
			return invalid()
		}
		target, err := url.Parse(response.ContinuationURL)
		if err != nil || target.Scheme != "https" || target.Host == "" || target.User != nil || target.Fragment != "" {
			return invalid()
		}
		return &response, nil
	case "custom_completion_required":
		var response PasswordlessCustomCompletionRequired
		var requirements map[string]json.RawMessage
		if json.Unmarshal(raw, &response) != nil || response.Continuation == "" || response.ExpiresIn <= 0 || response.ExpiresIn > 300 || json.Unmarshal(fields["requirements"], &requirements) != nil {
			return invalid()
		}
		for _, key := range []string{"registration", "consent", "scopes", "aether_terms_version", "aether_privacy_version"} {
			if requirements[key] == nil {
				return invalid()
			}
		}
		if string(requirements["registration"]) == "null" || string(requirements["consent"]) == "null" || response.Requirements.Scopes == nil {
			return invalid()
		}
		if response.Requirements.Registration && (response.Requirements.AetherTermsVersion == nil || response.Requirements.AetherPrivacyVersion == nil) {
			return invalid()
		}
		return &response, nil
	case "denied":
		var response PasswordlessDenied
		if json.Unmarshal(raw, &response) != nil || response.Error != "access_denied" || fields["state"] == nil {
			return invalid()
		}
		return &response, nil
	default:
		return invalid()
	}
}

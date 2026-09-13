package identity

import (
	"errors"
	"fmt"
	"github.com/aetherplatform/aether-go"
	"strings"
)

func sanitizeConfidentialError(err error, _ ...string) error {
	if err == nil {
		return nil
	}
	var original *aether.Error
	if !errors.As(err, &original) {
		return fmt.Errorf("identity: authentication request failed")
	}
	clean := *original
	clean.Details, clean.Cause = nil, nil
	clean.RequestID, clean.Retry.QuotaReset = "", ""
	clean.Message = "Identity authentication request failed"
	if !safeIdentityCode(clean.Code) {
		clean.Code = "request_failed"
	}
	return &clean
}

func safeIdentityCode(code string) bool {
	switch strings.ToUpper(code) {
	case "INVALID_REQUEST", "INVALID_PASSWORDLESS_REQUEST", "INVALID_CLIENT", "ORIGIN_NOT_ALLOWED",
		"CUSTOM_COMPLETION_NOT_ALLOWED", "RATE_LIMITED", "TEMPORARILY_UNAVAILABLE", "IDENTITY_ROLE_UNAVAILABLE",
		"UNSUPPORTED_GRANT_TYPE", "UNAUTHORIZED_CLIENT", "INVALID_TARGET", "INVALID_SCOPE", "TOKEN_ISSUANCE_UNAVAILABLE",
		"INVALID_GRANT", "INVALID_CODE", "CODE_ALREADY_USED", "CODE_EXPIRED", "INVALID_REDIRECT_URI", "INVALID_CODE_VERIFIER",
		"REVOCATION_BACKEND_UNAVAILABLE", "INVALID_TOKEN", "FEDERATION_SERVICE_UNAVAILABLE", "REQUEST_FAILED",
		"REQUEST_ABORTED", "REQUEST_TIMEOUT", "REQUEST_BUILD_FAILED", "NETWORK_ERROR", "RESPONSE_READ_FAILED",
		"INVALID_RESPONSE", "INVALID_TOKEN_RESPONSE":
		return true
	default:
		return false
	}
}

func validOperationStatus(operation string, status int) bool {
	for _, op := range allOperations {
		if op.Name == operation {
			for _, allowed := range op.SuccessStatuses {
				if status == allowed {
					return true
				}
			}
			return false
		}
	}
	return false
}

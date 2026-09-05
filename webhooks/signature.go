package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SignatureErrorCode string

const (
	SignatureMissing          SignatureErrorCode = "missing_signature"
	SignatureTimestampMissing SignatureErrorCode = "missing_timestamp"
	SignatureFormatInvalid    SignatureErrorCode = "invalid_signature_format"
	SignatureTimestampExpired SignatureErrorCode = "timestamp_expired"
	SignatureInvalid          SignatureErrorCode = "invalid_signature"
)

type SignatureError struct {
	Code SignatureErrorCode
}

func (err *SignatureError) Error() string {
	if err == nil {
		return ""
	}
	return string(err.Code)
}

type VerifySignatureOptions struct {
	Tolerance time.Duration
	Now       time.Time
}

type VerifiedSignature struct {
	Timestamp time.Time
}

func VerifySignature(payload []byte, signatureHeader string, secret []byte, options VerifySignatureOptions) (VerifiedSignature, error) {
	if len(secret) == 0 {
		return VerifiedSignature{}, fmt.Errorf("webhooks: secret must not be empty")
	}
	tolerance := options.Tolerance
	if tolerance == 0 {
		tolerance = 5 * time.Minute
	}
	if tolerance < 0 {
		return VerifiedSignature{}, fmt.Errorf("webhooks: signature tolerance must not be negative")
	}
	now := options.Now
	if now.IsZero() {
		now = time.Now()
	}
	timestamp, signatures, err := parseSignatureHeader(signatureHeader)
	if err != nil {
		return VerifiedSignature{}, err
	}
	signedAt := time.Unix(timestamp, 0)
	age := now.Sub(signedAt)
	if age < -tolerance || age > tolerance {
		return VerifiedSignature{}, &SignatureError{Code: SignatureTimestampExpired}
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = fmt.Fprintf(mac, "%d.", timestamp)
	_, _ = mac.Write(payload)
	expected := mac.Sum(nil)
	for _, signature := range signatures {
		if hmac.Equal(expected, signature) {
			return VerifiedSignature{Timestamp: signedAt}, nil
		}
	}
	return VerifiedSignature{}, &SignatureError{Code: SignatureInvalid}
}

func parseSignatureHeader(header string) (int64, [][]byte, error) {
	if header == "" {
		return 0, nil, &SignatureError{Code: SignatureMissing}
	}
	var timestamps []string
	var signatures [][]byte
	for _, component := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(component), "=")
		if len(parts) != 2 || parts[1] == "" {
			return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
		}
		switch parts[0] {
		case "t":
			timestamps = append(timestamps, parts[1])
		case "v1":
			if len(parts[1]) != sha256.Size*2 || strings.ToLower(parts[1]) != parts[1] {
				return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
			}
			digest, err := hex.DecodeString(parts[1])
			if err != nil {
				return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
			}
			signatures = append(signatures, digest)
		default:
			return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
		}
	}
	if len(timestamps) == 0 {
		return 0, nil, &SignatureError{Code: SignatureTimestampMissing}
	}
	if len(timestamps) != 1 {
		return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
	}
	if len(signatures) == 0 {
		return 0, nil, &SignatureError{Code: SignatureMissing}
	}
	if timestamps[0] != "0" && strings.HasPrefix(timestamps[0], "0") {
		return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
	}
	timestamp, err := strconv.ParseInt(timestamps[0], 10, 64)
	if err != nil || timestamp < 0 {
		return 0, nil, &SignatureError{Code: SignatureFormatInvalid}
	}
	return timestamp, signatures, nil
}

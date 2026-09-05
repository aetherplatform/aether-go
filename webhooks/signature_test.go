package webhooks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type signatureVector struct {
	Name       string `json:"name"`
	Payload    string `json:"payload"`
	Secret     string `json:"secret"`
	NowSeconds int64  `json:"now_seconds"`
	Header     string `json:"header"`
	Timestamp  int64  `json:"timestamp"`
	Error      string `json:"error"`
}

func TestSignatureVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Tolerance int64             `json:"default_tolerance_seconds"`
		Valid     []signatureVector `json:"valid"`
		Invalid   []signatureVector `json:"invalid"`
	}
	loadSignatureVectors(t, &document)
	for _, vector := range document.Valid {
		vector := vector
		t.Run(vector.Name, func(t *testing.T) {
			verified, err := VerifySignature([]byte(vector.Payload), vector.Header, []byte(vector.Secret), VerifySignatureOptions{
				Tolerance: time.Duration(document.Tolerance) * time.Second,
				Now:       time.Unix(vector.NowSeconds, 0),
			})
			if err != nil {
				t.Fatal(err)
			}
			if verified.Timestamp.Unix() != vector.Timestamp {
				t.Fatalf("timestamp=%d", verified.Timestamp.Unix())
			}
		})
	}
	for _, vector := range document.Invalid {
		vector := vector
		t.Run(vector.Name, func(t *testing.T) {
			_, err := VerifySignature([]byte(vector.Payload), vector.Header, []byte(vector.Secret), VerifySignatureOptions{
				Tolerance: time.Duration(document.Tolerance) * time.Second,
				Now:       time.Unix(vector.NowSeconds, 0),
			})
			var signatureErr *SignatureError
			if !errors.As(err, &signatureErr) || string(signatureErr.Code) != vector.Error {
				t.Fatalf("error=%v, want %s", err, vector.Error)
			}
		})
	}
}

func loadSignatureVectors(t *testing.T, target any) {
	t.Helper()
	paths := []string{
		filepath.Join("..", "..", "..", "contracts", "webhooks", "signatures", "v1", "vectors.json"),
		filepath.Join("..", "contracts", "webhooks", "signatures", "v1", "vectors.json"),
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
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

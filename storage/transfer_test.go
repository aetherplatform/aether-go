package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aetherplatform/aether-go"
)

func TestMultipartVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Cases []struct {
			Name        string    `json:"name"`
			Size        int64     `json:"size"`
			PartSize    int64     `json:"part_size"`
			PartNumbers []int64   `json:"part_numbers"`
			Ranges      [][]int64 `json:"ranges"`
			Valid       bool      `json:"valid"`
		} `json:"cases"`
	}
	loadTransferVectors(t, &document)
	for _, testCase := range document.Cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			parts := make([]MultipartPart, len(testCase.PartNumbers))
			for index, partNumber := range testCase.PartNumbers {
				parts[index] = MultipartPart{PartNumber: partNumber, UploadURL: "https://provider.useather.test/" + strconv.FormatInt(partNumber, 10)}
			}
			intent := UploadIntent{UploadStrategy: UploadStrategyMultipart, Multipart: &MultipartInstructions{PartSize: testCase.PartSize, Parts: parts}}
			_, ranges, err := validateMultipartIntent(intent, bytes.NewReader(make([]byte, max(testCase.Size, 0))), testCase.Size)
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid=%v error=%v", testCase.Valid, err)
			}
			if !testCase.Valid {
				return
			}
			for index, expected := range testCase.Ranges {
				if ranges[index].start != expected[0] || ranges[index].end != expected[1] {
					t.Fatalf("range %d=%+v, want %v", index, ranges[index], expected)
				}
			}
		})
	}
}

func TestMultipartUploadBoundsConcurrencyAndPreservesETags(t *testing.T) {
	t.Parallel()
	var active atomic.Int64
	var maximum atomic.Int64
	var mu sync.Mutex
	uploaded := make(map[string]string)
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("authorization") != "" || request.Header.Get("x-request-id") != "" || request.Header.Get("x-correlation-id") != "" {
			t.Errorf("Aether credential headers reached provider")
		}
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		body, _ := io.ReadAll(request.Body)
		mu.Lock()
		uploaded[request.URL.Path] = string(body)
		mu.Unlock()
		active.Add(-1)
		part := strings.TrimPrefix(request.URL.Path, "/part/")
		header := make(http.Header)
		header.Set("etag", `"etag-`+part+`"`)
		return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	intent := multipartIntent()
	parts, err := UploadMultipart(context.Background(), intent, bytes.NewReader([]byte("abcdefgh")), 8, MultipartUploadOptions{Concurrency: 2, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	if maximum.Load() != 2 {
		t.Fatalf("maximum concurrency=%d, want 2", maximum.Load())
	}
	if uploaded["/part/1"] != "abc" || uploaded["/part/2"] != "def" || uploaded["/part/3"] != "gh" {
		t.Fatalf("unexpected uploaded ranges: %#v", uploaded)
	}
	for index, part := range parts {
		wantNumber := int64(index + 1)
		wantETag := `"etag-` + strconv.FormatInt(wantNumber, 10) + `"`
		if part.PartNumber != wantNumber || part.ETag != wantETag {
			t.Fatalf("part %d=%+v", index, part)
		}
	}
}

func TestProviderRequestsRejectAetherHeaders(t *testing.T) {
	t.Parallel()
	err := UploadDirect(context.Background(), "https://provider.useather.test/upload", strings.NewReader("hello"), 5, http.Header{"Authorization": {"Bearer forbidden"}}, &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("provider request should not be sent")
		return nil, nil
	})})
	var transferErr *TransferError
	if !errors.As(err, &transferErr) || transferErr.Code != "invalid_upload_intent" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadWorkflowCompletesAndUsesSeparateProviderRequest(t *testing.T) {
	t.Parallel()
	var calls []string
	var mu sync.Mutex
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		mu.Lock()
		calls = append(calls, request.Method+" "+request.URL.String())
		mu.Unlock()
		if request.URL.Host == "provider.useather.test" {
			if request.Header.Get("authorization") != "" || request.Header.Get("x-request-id") != "" {
				t.Errorf("Aether headers reached provider: %#v", request.Header)
			}
			return transferResponse(http.StatusOK, "", nil), nil
		}
		switch request.URL.Path {
		case "/v1/storage/upload-intents":
			return transferResponse(http.StatusOK, singleIntentJSON(), nil), nil
		case "/v1/storage/uploads/upl_1/complete":
			return transferResponse(http.StatusOK, objectJSON("quarantined"), nil), nil
		case "/v1/storage/objects/obj_1":
			return transferResponse(http.StatusOK, objectJSON("available"), nil), nil
		default:
			return transferResponse(http.StatusNotFound, `{}`, nil), nil
		}
	})}
	client, err := NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: staticTokenProvider{}, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	waitOptions := WaitForObjectOptions{PollInterval: time.Millisecond, Timeout: time.Second}
	result, err := client.UploadObject(context.Background(), CreateUploadIntentRequest{
		NamespaceID:  "ns_1",
		LogicalKey:   "example.txt",
		Filename:     "example.txt",
		ContentType:  "text/plain",
		ExpectedSize: 5,
	}, bytes.NewReader([]byte("hello")), 5, UploadObjectOptions{HTTPClient: httpClient, WaitForAvailability: &waitOptions})
	if err != nil {
		t.Fatal(err)
	}
	if result.Object.Status != ObjectStatusAvailable {
		t.Fatalf("status=%s", result.Object.Status)
	}
	if len(calls) != 4 {
		t.Fatalf("calls=%v", calls)
	}
}

func TestUploadWorkflowPreservesCleanupFailure(t *testing.T) {
	t.Parallel()
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host == "provider.useather.test" {
			return transferResponse(http.StatusServiceUnavailable, "", nil), nil
		}
		switch request.URL.Path {
		case "/v1/storage/upload-intents":
			return transferResponse(http.StatusOK, singleIntentJSON(), nil), nil
		case "/v1/storage/uploads/upl_1/abort":
			return transferResponse(http.StatusBadRequest, `{"error":{"code":"cleanup_failed","message":"cleanup failed"}}`, nil), nil
		default:
			return transferResponse(http.StatusNotFound, `{}`, nil), nil
		}
	})}
	client, err := NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: staticTokenProvider{}, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.UploadObject(context.Background(), CreateUploadIntentRequest{NamespaceID: "ns_1", LogicalKey: "example.txt", Filename: "example.txt", ContentType: "text/plain", ExpectedSize: 5}, bytes.NewReader([]byte("hello")), 5, UploadObjectOptions{HTTPClient: httpClient})
	var workflowErr *UploadWorkflowError
	if !errors.As(err, &workflowErr) || workflowErr.Stage != UploadStageTransfer || workflowErr.CleanupError == nil {
		t.Fatalf("unexpected workflow error: %#v", err)
	}
}

func transferResponse(status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body))}
}

func multipartIntent() UploadIntent {
	return UploadIntent{
		UploadID:        "upl_1",
		ObjectID:        "obj_1",
		UploadStrategy:  UploadStrategyMultipart,
		RequiredHeaders: map[string]string{"content-type": "text/plain"},
		Multipart: &MultipartInstructions{
			PartSize: 3,
			Parts: []MultipartPart{
				{PartNumber: 3, UploadURL: "https://provider.useather.test/part/3"},
				{PartNumber: 1, UploadURL: "https://provider.useather.test/part/1"},
				{PartNumber: 2, UploadURL: "https://provider.useather.test/part/2"},
			},
		},
	}
}

func singleIntentJSON() string {
	return `{"upload_id":"upl_1","object_id":"obj_1","asset_id":null,"version_number":1,"upload_strategy":"single","upload_url":"https://provider.useather.test/upload","expires_at":"2026-09-05T12:00:00Z","required_headers":{"content-type":"text/plain"},"multipart":null}`
}

func objectJSON(status string) string {
	return `{"id":"obj_1","asset_id":null,"namespace_id":"ns_1","organization_id":"org_1","application_id":"app_1","account_mode":"sandbox","logical_key":"example.txt","version_number":1,"filename":"example.txt","content_type":"text/plain","detected_content_type":null,"expected_size":5,"actual_size":5,"sha256":null,"metadata":{},"visibility":"private","status":"` + status + `","created_at":"2026-09-05T12:00:00Z","available_at":null,"deleted_at":null}`
}

func loadTransferVectors(t *testing.T, target any) {
	t.Helper()
	paths := []string{
		filepath.Join("..", "..", "..", "contracts", "sdk", "v1", "storage-transfer-vectors.json"),
		filepath.Join("..", "contracts", "sdk", "v1", "storage-transfer-vectors.json"),
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
		t.Fatalf("read transfer vectors: %v", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

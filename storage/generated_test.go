package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type staticTokenProvider struct{}

func (staticTokenProvider) Token(context.Context) (aether.AccessToken, error) {
	return aether.AccessToken{Value: "sdk-token", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestGeneratedOperationsMatchPublicAllowlist(t *testing.T) {
	t.Parallel()
	operations := []transport.Operation{
		abortUploadOperation,
		completeUploadOperation,
		createAssetOperation,
		createNamespaceOperation,
		createObjectDownloadIntentOperation,
		createObjectDownloadIntentByKeyOperation,
		createUploadIntentOperation,
		deleteAssetOperation,
		deleteObjectOperation,
		getAssetOperation,
		getNamespaceOperation,
		getNamespaceUsageOperation,
		getObjectOperation,
		listAssetVersionsOperation,
		listNamespacesOperation,
		resolveObjectOperation,
		restoreObjectOperation,
		searchObjectsOperation,
	}
	if len(operations) != 18 {
		t.Fatalf("operations=%d, want 18", len(operations))
	}
	seen := make(map[string]struct{}, len(operations))
	for _, operation := range operations {
		if _, exists := seen[operation.Name]; exists {
			t.Fatalf("duplicate operation %s", operation.Name)
		}
		seen[operation.Name] = struct{}{}
		if strings.Contains(operation.Path, "provider-migrations") {
			t.Fatalf("operator-private route was generated: %s", operation.Path)
		}
		if !strings.HasPrefix(operation.Path, "/v1/storage/") {
			t.Fatalf("operation uses non-public path: %s", operation.Path)
		}
	}
}

func TestAllGeneratedMethodsIssueExpectedRequests(t *testing.T) {
	t.Parallel()
	type capturedRequest struct {
		method         string
		path           string
		query          string
		idempotencyKey string
	}
	var mu sync.Mutex
	var captured []capturedRequest
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		mu.Lock()
		captured = append(captured, capturedRequest{
			method:         request.Method,
			path:           request.URL.EscapedPath(),
			query:          request.URL.RawQuery,
			idempotencyKey: request.Header.Get("idempotency-key"),
		})
		mu.Unlock()
		body := `{}`
		if request.Method == http.MethodGet && (request.URL.Path == "/v1/storage/namespaces" || strings.HasSuffix(request.URL.Path, "/versions")) {
			body = `[]`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	client, err := NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: staticTokenProvider{}, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	mutation := []aether.RequestOption{aether.WithIdempotencyKey("idem")}
	version := int64(2)
	prefix := "folder/"
	status := ObjectStatusAvailable
	created := time.Date(2026, 9, 5, 12, 30, 0, 0, time.UTC)
	limit := int64(25)

	calls := []struct {
		name   string
		method string
		path   string
		query  string
		call   func() error
	}{
		{"abortUpload", http.MethodPost, "/v1/storage/uploads/upl_1/abort", "", func() error { _, err := client.AbortUpload(ctx, "upl_1", mutation...); return err }},
		{"completeUpload", http.MethodPost, "/v1/storage/uploads/upl_1/complete", "", func() error {
			_, err := client.CompleteUpload(ctx, "upl_1", CompleteUploadRequest{}, mutation...)
			return err
		}},
		{"createAsset", http.MethodPost, "/v1/storage/assets", "", func() error {
			_, err := client.CreateAsset(ctx, CreateAssetRequest{NamespaceID: "ns_1"}, mutation...)
			return err
		}},
		{"createNamespace", http.MethodPost, "/v1/storage/namespaces", "", func() error {
			_, err := client.CreateNamespace(ctx, CreateNamespaceRequest{Name: "media"}, mutation...)
			return err
		}},
		{"createObjectDownloadIntent", http.MethodGet, "/v1/storage/objects/obj_1/download-intent", "", func() error { _, err := client.CreateObjectDownloadIntent(ctx, "obj_1"); return err }},
		{"createObjectDownloadIntentByKey", http.MethodGet, "/v1/storage/objects/download-intent", "key=folder%2Fa.txt&namespace=media", func() error {
			_, err := client.CreateObjectDownloadIntentByKey(ctx, CreateObjectDownloadIntentByKeyParams{Namespace: "media", Key: "folder/a.txt"})
			return err
		}},
		{"createUploadIntent", http.MethodPost, "/v1/storage/upload-intents", "", func() error {
			_, err := client.CreateUploadIntent(ctx, CreateUploadIntentRequest{NamespaceID: "ns_1", LogicalKey: "a", Filename: "a", ContentType: "text/plain"}, mutation...)
			return err
		}},
		{"deleteAsset", http.MethodDelete, "/v1/storage/assets/ast_1", "", func() error { _, err := client.DeleteAsset(ctx, "ast_1", mutation...); return err }},
		{"deleteObject", http.MethodDelete, "/v1/storage/objects/obj_1", "", func() error { _, err := client.DeleteObject(ctx, "obj_1", mutation...); return err }},
		{"getAsset", http.MethodGet, "/v1/storage/assets/ast_1", "", func() error { _, err := client.GetAsset(ctx, "ast_1"); return err }},
		{"getNamespace", http.MethodGet, "/v1/storage/namespaces/ns_1", "", func() error { _, err := client.GetNamespace(ctx, "ns_1"); return err }},
		{"getNamespaceUsage", http.MethodGet, "/v1/storage/usage", "namespace_id=ns_1", func() error {
			_, err := client.GetNamespaceUsage(ctx, GetNamespaceUsageParams{NamespaceID: "ns_1"})
			return err
		}},
		{"getObject", http.MethodGet, "/v1/storage/objects/obj_1", "", func() error { _, err := client.GetObject(ctx, "obj_1"); return err }},
		{"listAssetVersions", http.MethodGet, "/v1/storage/assets/ast_1/versions", "", func() error { _, err := client.ListAssetVersions(ctx, "ast_1"); return err }},
		{"listNamespaces", http.MethodGet, "/v1/storage/namespaces", "", func() error { _, err := client.ListNamespaces(ctx); return err }},
		{"resolveObject", http.MethodGet, "/v1/storage/objects", "key=folder%2Fa.txt&namespace_id=ns_1&version=2", func() error {
			_, err := client.ResolveObject(ctx, ResolveObjectParams{NamespaceID: "ns_1", Key: "folder/a.txt", Version: &version})
			return err
		}},
		{"restoreObject", http.MethodPost, "/v1/storage/objects/obj_1/restore", "", func() error { _, err := client.RestoreObject(ctx, "obj_1", mutation...); return err }},
		{"searchObjects", http.MethodGet, "/v1/storage/objects/search", "created_after=2026-09-05T12%3A30%3A00Z&limit=25&namespace_id=ns_1&prefix=folder%2F&status=available", func() error {
			_, err := client.SearchObjects(ctx, SearchObjectsParams{NamespaceID: "ns_1", Prefix: &prefix, Status: &status, CreatedAfter: &created, Limit: &limit})
			return err
		}},
	}

	for index, call := range calls {
		if err := call.call(); err != nil {
			t.Fatalf("%s: %v", call.name, err)
		}
		request := captured[index]
		if request.method != call.method || request.path != call.path || request.query != call.query {
			t.Fatalf("%s request=%s %s?%s", call.name, request.method, request.path, request.query)
		}
		if call.method != http.MethodGet && request.idempotencyKey != "idem" {
			t.Fatalf("%s omitted idempotency key", call.name)
		}
	}
	if len(captured) != 18 {
		t.Fatalf("requests=%d, want 18", len(captured))
	}
}

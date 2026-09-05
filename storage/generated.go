// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.
// Source: contracts/openapi/v1/storage.yaml

package storage

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type AbortUploadResult struct {
	Status   string   `json:"status"`
	UploadID UploadID `json:"upload_id"`
}

type Asset struct {
	AccountMode    AccountMode    `json:"account_mode"`
	ApplicationID  string         `json:"application_id"`
	CreatedAt      time.Time      `json:"created_at"`
	EntityID       *string        `json:"entity_id"`
	EntityType     *string        `json:"entity_type"`
	ID             AssetID        `json:"id"`
	Metadata       Metadata       `json:"metadata"`
	NamespaceID    NamespaceID    `json:"namespace_id"`
	OrganizationID string         `json:"organization_id"`
	Status         ResourceStatus `json:"status"`
}

type AssetID string

type CompleteUploadRequest struct {
	Parts  *[]CompletedPart `json:"parts,omitempty"`
	SHA256 *string          `json:"sha256,omitempty"`
}

type CompletedPart struct {
	ETag       string `json:"etag"`
	PartNumber int64  `json:"part_number"`
}

type CreateAssetRequest struct {
	EntityID    *string     `json:"entity_id,omitempty"`
	EntityType  *string     `json:"entity_type,omitempty"`
	Metadata    *Metadata   `json:"metadata,omitempty"`
	NamespaceID NamespaceID `json:"namespace_id"`
}

type CreateNamespaceRequest struct {
	Name   string           `json:"name"`
	Policy *NamespacePolicy `json:"policy,omitempty"`
}

type CreateUploadIntentRequest struct {
	AssetID      *AssetID    `json:"asset_id,omitempty"`
	ClientSHA256 *string     `json:"client_sha256,omitempty"`
	ContentType  string      `json:"content_type"`
	ExpectedSize int64       `json:"expected_size"`
	Filename     string      `json:"filename"`
	LogicalKey   LogicalKey  `json:"logical_key"`
	Metadata     *Metadata   `json:"metadata,omitempty"`
	NamespaceID  NamespaceID `json:"namespace_id"`
	Visibility   *Visibility `json:"visibility,omitempty"`
}

type DeleteAssetResult struct {
	AssetID      AssetID `json:"asset_id"`
	DeletedCount int64   `json:"deleted_count"`
	Status       string  `json:"status"`
}

type DownloadIntent struct {
	Delivery    Delivery   `json:"delivery"`
	DownloadURL string     `json:"download_url"`
	ExpiresAt   *time.Time `json:"expires_at"`
	ObjectID    ObjectID   `json:"object_id"`
}

type KeyedDownloadIntent struct {
	Checksum    *string    `json:"checksum"`
	ContentType string     `json:"content_type"`
	Delivery    Delivery   `json:"delivery"`
	DownloadURL string     `json:"download_url"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Filename    string     `json:"filename"`
	ObjectID    ObjectID   `json:"object_id"`
	Size        *int64     `json:"size"`
}

type LifecyclePolicy struct {
	DeletedRetentionSeconds *int64 `json:"deleted_retention_seconds,omitempty"`
	ExpireAfterSeconds      *int64 `json:"expire_after_seconds,omitempty"`
	KeepLatestVersions      *int64 `json:"keep_latest_versions,omitempty"`
}

type LogicalKey string

type Metadata map[string]any

type MultipartInstructions struct {
	PartSize         int64           `json:"part_size"`
	Parts            []MultipartPart `json:"parts"`
	ProviderUploadID string          `json:"provider_upload_id"`
}

type MultipartPart struct {
	PartNumber int64  `json:"part_number"`
	UploadURL  string `json:"upload_url"`
}

type Namespace struct {
	AccountMode    AccountMode     `json:"account_mode"`
	ApplicationID  string          `json:"application_id"`
	CreatedAt      time.Time       `json:"created_at"`
	ID             NamespaceID     `json:"id"`
	Name           string          `json:"name"`
	OrganizationID string          `json:"organization_id"`
	Policy         NamespacePolicy `json:"policy"`
	Status         ResourceStatus  `json:"status"`
}

type NamespaceID string

type NamespacePolicy struct {
	Lifecycle *LifecyclePolicy `json:"lifecycle,omitempty"`
	Quota     *QuotaPolicy     `json:"quota,omitempty"`
}

type Object struct {
	AccountMode         AccountMode  `json:"account_mode"`
	ActualSize          *int64       `json:"actual_size"`
	ApplicationID       string       `json:"application_id"`
	AssetID             *AssetID     `json:"asset_id"`
	AvailableAt         *time.Time   `json:"available_at"`
	ContentType         string       `json:"content_type"`
	CreatedAt           time.Time    `json:"created_at"`
	DeletedAt           *time.Time   `json:"deleted_at"`
	DetectedContentType *string      `json:"detected_content_type"`
	ExpectedSize        int64        `json:"expected_size"`
	Filename            string       `json:"filename"`
	ID                  ObjectID     `json:"id"`
	LogicalKey          LogicalKey   `json:"logical_key"`
	Metadata            Metadata     `json:"metadata"`
	NamespaceID         NamespaceID  `json:"namespace_id"`
	OrganizationID      string       `json:"organization_id"`
	SHA256              *string      `json:"sha256"`
	Status              ObjectStatus `json:"status"`
	VersionNumber       int64        `json:"version_number"`
	Visibility          Visibility   `json:"visibility"`
}

type ObjectID string

type ObjectSearchResponse struct {
	NextCursor *string  `json:"next_cursor"`
	Objects    []Object `json:"objects"`
}

type ObjectStatus string

const (
	ObjectStatusPending     ObjectStatus = "pending"
	ObjectStatusUploading   ObjectStatus = "uploading"
	ObjectStatusAvailable   ObjectStatus = "available"
	ObjectStatusQuarantined ObjectStatus = "quarantined"
	ObjectStatusProcessing  ObjectStatus = "processing"
	ObjectStatusDeleted     ObjectStatus = "deleted"
)

type QuotaPolicy struct {
	MaxBytes   *int64 `json:"max_bytes,omitempty"`
	MaxObjects *int64 `json:"max_objects,omitempty"`
}

type UploadID string

type UploadIntent struct {
	AssetID         *AssetID               `json:"asset_id"`
	ExpiresAt       time.Time              `json:"expires_at"`
	Multipart       *MultipartInstructions `json:"multipart"`
	ObjectID        ObjectID               `json:"object_id"`
	RequiredHeaders map[string]string      `json:"required_headers"`
	UploadID        UploadID               `json:"upload_id"`
	UploadStrategy  UploadStrategy         `json:"upload_strategy"`
	UploadURL       *string                `json:"upload_url"`
	VersionNumber   int64                  `json:"version_number"`
}

type Usage struct {
	AccountMode     AccountMode `json:"account_mode"`
	ApplicationID   string      `json:"application_id"`
	NamespaceID     NamespaceID `json:"namespace_id"`
	OrganizationID  string      `json:"organization_id"`
	QuotaMaxBytes   *int64      `json:"quota_max_bytes"`
	QuotaMaxObjects *int64      `json:"quota_max_objects"`
	ReservedBytes   int64       `json:"reserved_bytes"`
	ReservedObjects int64       `json:"reserved_objects"`
	StoredBytes     int64       `json:"stored_bytes"`
	StoredObjects   int64       `json:"stored_objects"`
	TotalBytes      int64       `json:"total_bytes"`
	TotalObjects    int64       `json:"total_objects"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type AccountMode string

const (
	AccountModeSandbox AccountMode = "sandbox"
	AccountModeLive    AccountMode = "live"
)

type Delivery string

const (
	DeliveryCDN            Delivery = "cdn"
	DeliverySignedProvider Delivery = "signed_provider"
)

type ResourceStatus string

const (
	ResourceStatusActive   ResourceStatus = "active"
	ResourceStatusDisabled ResourceStatus = "disabled"
)

type UploadStrategy string

const (
	UploadStrategySingle    UploadStrategy = "single"
	UploadStrategyMultipart UploadStrategy = "multipart"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type CreateObjectDownloadIntentByKeyParams struct {
	Key       LogicalKey
	Namespace string
}

type GetNamespaceUsageParams struct {
	NamespaceID NamespaceID
}

type ResolveObjectParams struct {
	Key         LogicalKey
	NamespaceID NamespaceID
	Version     *int64
}

type SearchObjectsParams struct {
	AssetID       *AssetID
	ContentType   *string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Cursor        *string
	EntityID      *string
	EntityType    *string
	Limit         *int64
	NamespaceID   NamespaceID
	Prefix        *string
	Status        *ObjectStatus
}

var (
	abortUploadOperation                     = transport.Operation{Name: "abortUpload", Method: http.MethodPost, Path: "/v1/storage/uploads/{upload_id}/abort", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	completeUploadOperation                  = transport.Operation{Name: "completeUpload", Method: http.MethodPost, Path: "/v1/storage/uploads/{upload_id}/complete", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	createAssetOperation                     = transport.Operation{Name: "createAsset", Method: http.MethodPost, Path: "/v1/storage/assets", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	createNamespaceOperation                 = transport.Operation{Name: "createNamespace", Method: http.MethodPost, Path: "/v1/storage/namespaces", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	createObjectDownloadIntentOperation      = transport.Operation{Name: "createObjectDownloadIntent", Method: http.MethodGet, Path: "/v1/storage/objects/{object_id}/download-intent", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	createObjectDownloadIntentByKeyOperation = transport.Operation{Name: "createObjectDownloadIntentByKey", Method: http.MethodGet, Path: "/v1/storage/objects/download-intent", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	createUploadIntentOperation              = transport.Operation{Name: "createUploadIntent", Method: http.MethodPost, Path: "/v1/storage/upload-intents", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	deleteAssetOperation                     = transport.Operation{Name: "deleteAsset", Method: http.MethodDelete, Path: "/v1/storage/assets/{asset_id}", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	deleteObjectOperation                    = transport.Operation{Name: "deleteObject", Method: http.MethodDelete, Path: "/v1/storage/objects/{object_id}", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	getAssetOperation                        = transport.Operation{Name: "getAsset", Method: http.MethodGet, Path: "/v1/storage/assets/{asset_id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getNamespaceOperation                    = transport.Operation{Name: "getNamespace", Method: http.MethodGet, Path: "/v1/storage/namespaces/{namespace_id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getNamespaceUsageOperation               = transport.Operation{Name: "getNamespaceUsage", Method: http.MethodGet, Path: "/v1/storage/usage", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getObjectOperation                       = transport.Operation{Name: "getObject", Method: http.MethodGet, Path: "/v1/storage/objects/{object_id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listAssetVersionsOperation               = transport.Operation{Name: "listAssetVersions", Method: http.MethodGet, Path: "/v1/storage/assets/{asset_id}/versions", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listNamespacesOperation                  = transport.Operation{Name: "listNamespaces", Method: http.MethodGet, Path: "/v1/storage/namespaces", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	resolveObjectOperation                   = transport.Operation{Name: "resolveObject", Method: http.MethodGet, Path: "/v1/storage/objects", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	restoreObjectOperation                   = transport.Operation{Name: "restoreObject", Method: http.MethodPost, Path: "/v1/storage/objects/{object_id}/restore", Idempotency: transport.IdempotencyRequired, SuccessStatuses: []int{200}}
	searchObjectsOperation                   = transport.Operation{Name: "searchObjects", Method: http.MethodGet, Path: "/v1/storage/objects/search", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
)

var allOperations = []transport.Operation{
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

func (client *Client) AbortUpload(ctx context.Context, uploadID UploadID, options ...aether.RequestOption) (*AbortUploadResult, error) {
	var query url.Values
	path := map[string]string{
		"upload_id": string(uploadID),
	}
	var response AbortUploadResult
	err := client.execute(ctx, abortUploadOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CompleteUpload(ctx context.Context, uploadID UploadID, request CompleteUploadRequest, options ...aether.RequestOption) (*Object, error) {
	var query url.Values
	path := map[string]string{
		"upload_id": string(uploadID),
	}
	var response Object
	err := client.execute(ctx, completeUploadOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateAsset(ctx context.Context, request CreateAssetRequest, options ...aether.RequestOption) (*Asset, error) {
	var query url.Values
	var path map[string]string
	var response Asset
	err := client.execute(ctx, createAssetOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateNamespace(ctx context.Context, request CreateNamespaceRequest, options ...aether.RequestOption) (*Namespace, error) {
	var query url.Values
	var path map[string]string
	var response Namespace
	err := client.execute(ctx, createNamespaceOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateObjectDownloadIntent(ctx context.Context, objectID ObjectID, options ...aether.RequestOption) (*DownloadIntent, error) {
	var query url.Values
	path := map[string]string{
		"object_id": string(objectID),
	}
	var response DownloadIntent
	err := client.execute(ctx, createObjectDownloadIntentOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CreateObjectDownloadIntentByKey(ctx context.Context, params CreateObjectDownloadIntentByKeyParams, options ...aether.RequestOption) (*KeyedDownloadIntent, error) {
	query := make(url.Values)
	query.Set("key", string(params.Key))
	query.Set("namespace", string(params.Namespace))
	var path map[string]string
	var response KeyedDownloadIntent
	err := client.execute(ctx, createObjectDownloadIntentByKeyOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CreateUploadIntent(ctx context.Context, request CreateUploadIntentRequest, options ...aether.RequestOption) (*UploadIntent, error) {
	var query url.Values
	var path map[string]string
	var response UploadIntent
	err := client.execute(ctx, createUploadIntentOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) DeleteAsset(ctx context.Context, assetID AssetID, options ...aether.RequestOption) (*DeleteAssetResult, error) {
	var query url.Values
	path := map[string]string{
		"asset_id": string(assetID),
	}
	var response DeleteAssetResult
	err := client.execute(ctx, deleteAssetOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) DeleteObject(ctx context.Context, objectID ObjectID, options ...aether.RequestOption) (*Object, error) {
	var query url.Values
	path := map[string]string{
		"object_id": string(objectID),
	}
	var response Object
	err := client.execute(ctx, deleteObjectOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetAsset(ctx context.Context, assetID AssetID, options ...aether.RequestOption) (*Asset, error) {
	var query url.Values
	path := map[string]string{
		"asset_id": string(assetID),
	}
	var response Asset
	err := client.execute(ctx, getAssetOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetNamespace(ctx context.Context, namespaceID NamespaceID, options ...aether.RequestOption) (*Namespace, error) {
	var query url.Values
	path := map[string]string{
		"namespace_id": string(namespaceID),
	}
	var response Namespace
	err := client.execute(ctx, getNamespaceOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetNamespaceUsage(ctx context.Context, params GetNamespaceUsageParams, options ...aether.RequestOption) (*Usage, error) {
	query := make(url.Values)
	query.Set("namespace_id", string(params.NamespaceID))
	var path map[string]string
	var response Usage
	err := client.execute(ctx, getNamespaceUsageOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetObject(ctx context.Context, objectID ObjectID, options ...aether.RequestOption) (*Object, error) {
	var query url.Values
	path := map[string]string{
		"object_id": string(objectID),
	}
	var response Object
	err := client.execute(ctx, getObjectOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListAssetVersions(ctx context.Context, assetID AssetID, options ...aether.RequestOption) ([]Object, error) {
	var query url.Values
	path := map[string]string{
		"asset_id": string(assetID),
	}
	var response []Object
	err := client.execute(ctx, listAssetVersionsOperation, path, query, nil, &response, options...)
	return response, err
}

func (client *Client) ListNamespaces(ctx context.Context, options ...aether.RequestOption) ([]Namespace, error) {
	var query url.Values
	var path map[string]string
	var response []Namespace
	err := client.execute(ctx, listNamespacesOperation, path, query, nil, &response, options...)
	return response, err
}

func (client *Client) ResolveObject(ctx context.Context, params ResolveObjectParams, options ...aether.RequestOption) (*Object, error) {
	query := make(url.Values)
	query.Set("key", string(params.Key))
	query.Set("namespace_id", string(params.NamespaceID))
	if params.Version != nil {
		query.Set("version", strconv.FormatInt(int64(*params.Version), 10))
	}
	var path map[string]string
	var response Object
	err := client.execute(ctx, resolveObjectOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) RestoreObject(ctx context.Context, objectID ObjectID, options ...aether.RequestOption) (*Object, error) {
	var query url.Values
	path := map[string]string{
		"object_id": string(objectID),
	}
	var response Object
	err := client.execute(ctx, restoreObjectOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) SearchObjects(ctx context.Context, params SearchObjectsParams, options ...aether.RequestOption) (*ObjectSearchResponse, error) {
	query := make(url.Values)
	if params.AssetID != nil {
		query.Set("asset_id", string(*params.AssetID))
	}
	if params.ContentType != nil {
		query.Set("content_type", string(*params.ContentType))
	}
	if params.CreatedAfter != nil {
		query.Set("created_after", params.CreatedAfter.Format(time.RFC3339))
	}
	if params.CreatedBefore != nil {
		query.Set("created_before", params.CreatedBefore.Format(time.RFC3339))
	}
	if params.Cursor != nil {
		query.Set("cursor", string(*params.Cursor))
	}
	if params.EntityID != nil {
		query.Set("entity_id", string(*params.EntityID))
	}
	if params.EntityType != nil {
		query.Set("entity_type", string(*params.EntityType))
	}
	if params.Limit != nil {
		query.Set("limit", strconv.FormatInt(int64(*params.Limit), 10))
	}
	query.Set("namespace_id", string(params.NamespaceID))
	if params.Prefix != nil {
		query.Set("prefix", string(*params.Prefix))
	}
	if params.Status != nil {
		query.Set("status", string(*params.Status))
	}
	var path map[string]string
	var response ObjectSearchResponse
	err := client.execute(ctx, searchObjectsOperation, path, query, nil, &response, options...)
	return &response, err
}
func setString(query url.Values, name string, value *string) {
	if value != nil {
		query.Set(name, *value)
	}
}
func setInteger(query url.Values, name string, value *int64) {
	if value != nil {
		query.Set(name, strconv.FormatInt(*value, 10))
	}
}
func setTime(query url.Values, name string, value *time.Time) {
	if value != nil {
		query.Set(name, value.Format(time.RFC3339))
	}
}

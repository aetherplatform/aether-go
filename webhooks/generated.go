// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.
// Source: contracts/openapi/v1/webhooks.yaml

package webhooks

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type ActionResult map[string]any

type BulkReplayRequest struct {
	From           *time.Time `json:"from,omitempty"`
	Limit          *int64     `json:"limit,omitempty"`
	SubscriptionID *string    `json:"subscription_id,omitempty"`
	To             *time.Time `json:"to,omitempty"`
}

type Delivery struct {
	Attempt        int64     `json:"attempt"`
	AttemptedAt    time.Time `json:"attempted_at"`
	DurationMs     *int64    `json:"duration_ms,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	EventID        string    `json:"event_id"`
	ID             string    `json:"id"`
	StatusCode     *int64    `json:"status_code,omitempty"`
	SubscriptionID string    `json:"subscription_id"`
	Success        bool      `json:"success"`
	URL            *string   `json:"url,omitempty"`
}

type DeliveryList struct {
	Data  []Delivery `json:"data"`
	Total int64      `json:"total"`
}

type HealthEnvelope struct {
	Data struct {
		Health         *map[string]any `json:"health"`
		SubscriptionID string          `json:"subscription_id"`
	} `json:"data"`
}

type InboundEndpoint struct {
	AllowedEventTypes *[]string      `json:"allowed_event_types,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	Description       *string        `json:"description,omitempty"`
	ID                string         `json:"id"`
	Secret            *string        `json:"secret,omitempty"`
	SecretRotatedAt   *time.Time     `json:"secret_rotated_at,omitempty"`
	Status            ResourceStatus `json:"status"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type InboundEndpointEnvelope struct {
	Data InboundEndpoint `json:"data"`
}

type InboundEndpointList struct {
	Data []InboundEndpoint `json:"data"`
}

type InboundEndpointWrite struct {
	AllowedEventTypes *[]string       `json:"allowed_event_types,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Status            *ResourceStatus `json:"status,omitempty"`
}

type JsonObject map[string]any

type PublishEventRequest struct {
	APIVersion     *string    `json:"api_version,omitempty"`
	Data           JsonObject `json:"data"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	Type           string     `json:"type"`
}

type PublishedEvent struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
}

type Subscription struct {
	AccountMode         *AccountMode   `json:"account_mode,omitempty"`
	APIVersion          *string        `json:"api_version,omitempty"`
	ConsecutiveFailures *int64         `json:"consecutive_failures,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	Description         *string        `json:"description,omitempty"`
	EnabledEvents       *[]string      `json:"enabled_events,omitempty"`
	Events              *[]string      `json:"events,omitempty"`
	ID                  string         `json:"id"`
	IsActive            *bool          `json:"is_active,omitempty"`
	LastFailureAt       *time.Time     `json:"last_failure_at,omitempty"`
	LastSuccessAt       *time.Time     `json:"last_success_at,omitempty"`
	Metadata            *JsonObject    `json:"metadata,omitempty"`
	PayloadVersion      *int64         `json:"payload_version,omitempty"`
	RetryPolicy         *JsonObject    `json:"retry_policy,omitempty"`
	Status              ResourceStatus `json:"status"`
	TransformSpec       *JsonObject    `json:"transform_spec,omitempty"`
	Transport           *string        `json:"transport,omitempty"`
	UpdatedAt           time.Time      `json:"updated_at"`
	URL                 *string        `json:"url,omitempty"`
}

type SubscriptionEnvelope struct {
	Data Subscription `json:"data"`
}

type SubscriptionList struct {
	Data  []Subscription `json:"data"`
	Total int64          `json:"total"`
}

type SubscriptionSecretEnvelope struct {
	Data   Subscription `json:"data"`
	Secret string       `json:"secret"`
}

type SubscriptionWrite struct {
	APIVersion     *string         `json:"api_version,omitempty"`
	Description    *string         `json:"description,omitempty"`
	EnabledEvents  *[]string       `json:"enabled_events,omitempty"`
	Events         *[]string       `json:"events,omitempty"`
	IsActive       *bool           `json:"is_active,omitempty"`
	Metadata       *JsonObject     `json:"metadata,omitempty"`
	PayloadVersion *int64          `json:"payload_version,omitempty"`
	RetryPolicy    *JsonObject     `json:"retry_policy,omitempty"`
	Status         *ResourceStatus `json:"status,omitempty"`
	TransformSpec  *JsonObject     `json:"transform_spec,omitempty"`
	Transport      *string         `json:"transport,omitempty"`
	URL            *string         `json:"url,omitempty"`
}

type AccountMode string

const (
	AccountModeSandbox AccountMode = "sandbox"
	AccountModeLive    AccountMode = "live"
)

type ResourceStatus string

const (
	ResourceStatusActive   ResourceStatus = "active"
	ResourceStatusDisabled ResourceStatus = "disabled"
)

type ListWebhookDeliveriesParams struct {
	Limit *int64
}

type ListWebhookSubscriptionsParams struct {
	Include *string
	Limit   *int64
}

var (
	bulkReplayWebhookDeliveriesOperation  = transport.Operation{Name: "bulkReplayWebhookDeliveries", Method: http.MethodPost, Path: "/v1/webhooks/bulk-replay", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{202}}
	createInboundEndpointOperation        = transport.Operation{Name: "createInboundEndpoint", Method: http.MethodPost, Path: "/v1/webhooks/inbound-endpoints", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{201}}
	createWebhookSubscriptionOperation    = transport.Operation{Name: "createWebhookSubscription", Method: http.MethodPost, Path: "/v1/webhooks/subscriptions", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{201}}
	deleteWebhookSubscriptionOperation    = transport.Operation{Name: "deleteWebhookSubscription", Method: http.MethodDelete, Path: "/v1/webhooks/subscriptions/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	getInboundEndpointOperation           = transport.Operation{Name: "getInboundEndpoint", Method: http.MethodGet, Path: "/v1/webhooks/inbound-endpoints/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getWebhookSubscriptionOperation       = transport.Operation{Name: "getWebhookSubscription", Method: http.MethodGet, Path: "/v1/webhooks/subscriptions/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getWebhookSubscriptionHealthOperation = transport.Operation{Name: "getWebhookSubscriptionHealth", Method: http.MethodGet, Path: "/v1/webhooks/subscriptions/{id}/health", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listInboundEndpointsOperation         = transport.Operation{Name: "listInboundEndpoints", Method: http.MethodGet, Path: "/v1/webhooks/inbound-endpoints", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listWebhookDeliveriesOperation        = transport.Operation{Name: "listWebhookDeliveries", Method: http.MethodGet, Path: "/v1/webhooks/subscriptions/{subscription_id}/deliveries", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listWebhookSubscriptionsOperation     = transport.Operation{Name: "listWebhookSubscriptions", Method: http.MethodGet, Path: "/v1/webhooks/subscriptions", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	patchWebhookSubscriptionOperation     = transport.Operation{Name: "patchWebhookSubscription", Method: http.MethodPatch, Path: "/v1/webhooks/subscriptions/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	publishWebhookEventOperation          = transport.Operation{Name: "publishWebhookEvent", Method: http.MethodPost, Path: "/v1/webhooks/events", Idempotency: transport.IdempotencyRequestField, SuccessStatuses: []int{202}, RequestRetrySafe: func(body any) bool {
		request, ok := body.(PublishEventRequest)
		return ok && request.IdempotencyKey != nil && strings.TrimSpace(string(*request.IdempotencyKey)) != ""
	}}
	replaceWebhookSubscriptionOperation      = transport.Operation{Name: "replaceWebhookSubscription", Method: http.MethodPut, Path: "/v1/webhooks/subscriptions/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	replayInboundEventOperation              = transport.Operation{Name: "replayInboundEvent", Method: http.MethodPost, Path: "/v1/webhooks/inbound-events/{id}/replay", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200, 202}}
	replayWebhookDeliveryOperation           = transport.Operation{Name: "replayWebhookDelivery", Method: http.MethodPost, Path: "/v1/webhooks/deliveries/{id}/replay", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	rotateInboundEndpointSecretOperation     = transport.Operation{Name: "rotateInboundEndpointSecret", Method: http.MethodPost, Path: "/v1/webhooks/inbound-endpoints/{id}/rotate-secret", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	rotateWebhookSubscriptionSecretOperation = transport.Operation{Name: "rotateWebhookSubscriptionSecret", Method: http.MethodPost, Path: "/v1/webhooks/subscriptions/{id}/rotate-secret", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	setWebhookSubscriptionStatusOperation    = transport.Operation{Name: "setWebhookSubscriptionStatus", Method: http.MethodPut, Path: "/v1/webhooks/subscriptions/{id}/status", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	updateInboundEndpointOperation           = transport.Operation{Name: "updateInboundEndpoint", Method: http.MethodPatch, Path: "/v1/webhooks/inbound-endpoints/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
)

var allOperations = []transport.Operation{
	bulkReplayWebhookDeliveriesOperation,
	createInboundEndpointOperation,
	createWebhookSubscriptionOperation,
	deleteWebhookSubscriptionOperation,
	getInboundEndpointOperation,
	getWebhookSubscriptionOperation,
	getWebhookSubscriptionHealthOperation,
	listInboundEndpointsOperation,
	listWebhookDeliveriesOperation,
	listWebhookSubscriptionsOperation,
	patchWebhookSubscriptionOperation,
	publishWebhookEventOperation,
	replaceWebhookSubscriptionOperation,
	replayInboundEventOperation,
	replayWebhookDeliveryOperation,
	rotateInboundEndpointSecretOperation,
	rotateWebhookSubscriptionSecretOperation,
	setWebhookSubscriptionStatusOperation,
	updateInboundEndpointOperation,
}

func (client *Client) BulkReplayWebhookDeliveries(ctx context.Context, request BulkReplayRequest, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	var path map[string]string
	var response JsonObject
	err := client.execute(ctx, bulkReplayWebhookDeliveriesOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateInboundEndpoint(ctx context.Context, request InboundEndpointWrite, options ...aether.RequestOption) (*InboundEndpointEnvelope, error) {
	var query url.Values
	var path map[string]string
	var response InboundEndpointEnvelope
	err := client.execute(ctx, createInboundEndpointOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateWebhookSubscription(ctx context.Context, request SubscriptionWrite, options ...aether.RequestOption) (*SubscriptionSecretEnvelope, error) {
	var query url.Values
	var path map[string]string
	var response SubscriptionSecretEnvelope
	err := client.execute(ctx, createWebhookSubscriptionOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) DeleteWebhookSubscription(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, deleteWebhookSubscriptionOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetInboundEndpoint(ctx context.Context, id string, options ...aether.RequestOption) (*InboundEndpointEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response InboundEndpointEnvelope
	err := client.execute(ctx, getInboundEndpointOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetWebhookSubscription(ctx context.Context, id string, options ...aether.RequestOption) (*SubscriptionEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response SubscriptionEnvelope
	err := client.execute(ctx, getWebhookSubscriptionOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetWebhookSubscriptionHealth(ctx context.Context, id string, options ...aether.RequestOption) (*HealthEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response HealthEnvelope
	err := client.execute(ctx, getWebhookSubscriptionHealthOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListInboundEndpoints(ctx context.Context, options ...aether.RequestOption) (*InboundEndpointList, error) {
	var query url.Values
	var path map[string]string
	var response InboundEndpointList
	err := client.execute(ctx, listInboundEndpointsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListWebhookDeliveries(ctx context.Context, subscriptionID string, params ListWebhookDeliveriesParams, options ...aether.RequestOption) (*DeliveryList, error) {
	query := make(url.Values)
	if params.Limit != nil {
		query.Set("limit", strconv.FormatInt(int64(*params.Limit), 10))
	}
	path := map[string]string{
		"subscription_id": string(subscriptionID),
	}
	var response DeliveryList
	err := client.execute(ctx, listWebhookDeliveriesOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListWebhookSubscriptions(ctx context.Context, params ListWebhookSubscriptionsParams, options ...aether.RequestOption) (*SubscriptionList, error) {
	query := make(url.Values)
	if params.Include != nil {
		query.Set("include", string(*params.Include))
	}
	if params.Limit != nil {
		query.Set("limit", strconv.FormatInt(int64(*params.Limit), 10))
	}
	var path map[string]string
	var response SubscriptionList
	err := client.execute(ctx, listWebhookSubscriptionsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) PatchWebhookSubscription(ctx context.Context, id string, request SubscriptionWrite, options ...aether.RequestOption) (*SubscriptionEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response SubscriptionEnvelope
	err := client.execute(ctx, patchWebhookSubscriptionOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) PublishWebhookEvent(ctx context.Context, request PublishEventRequest, options ...aether.RequestOption) (*PublishedEvent, error) {
	var query url.Values
	var path map[string]string
	var response PublishedEvent
	err := client.execute(ctx, publishWebhookEventOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) ReplaceWebhookSubscription(ctx context.Context, id string, request SubscriptionWrite, options ...aether.RequestOption) (*SubscriptionEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response SubscriptionEnvelope
	err := client.execute(ctx, replaceWebhookSubscriptionOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) ReplayInboundEvent(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, replayInboundEventOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ReplayWebhookDelivery(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, replayWebhookDeliveryOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) RotateInboundEndpointSecret(ctx context.Context, id string, options ...aether.RequestOption) (*InboundEndpointEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response InboundEndpointEnvelope
	err := client.execute(ctx, rotateInboundEndpointSecretOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) RotateWebhookSubscriptionSecret(ctx context.Context, id string, options ...aether.RequestOption) (*SubscriptionSecretEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response SubscriptionSecretEnvelope
	err := client.execute(ctx, rotateWebhookSubscriptionSecretOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) SetWebhookSubscriptionStatus(ctx context.Context, id string, request struct {
	Enabled  *bool `json:"enabled,omitempty"`
	IsActive *bool `json:"is_active,omitempty"`
}, options ...aether.RequestOption) (*SubscriptionEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response SubscriptionEnvelope
	err := client.execute(ctx, setWebhookSubscriptionStatusOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) UpdateInboundEndpoint(ctx context.Context, id string, request InboundEndpointWrite, options ...aether.RequestOption) (*InboundEndpointEnvelope, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response InboundEndpointEnvelope
	err := client.execute(ctx, updateInboundEndpointOperation, path, query, request, &response, options...)
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

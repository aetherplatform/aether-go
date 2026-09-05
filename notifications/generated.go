// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.
// Source: contracts/openapi/v1/notifications.yaml

package notifications

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type ActionResult map[string]any

type Analytics map[string]any

type EmailConfiguration map[string]any

type JsonObject map[string]any

type QueuedNotification struct {
	Message        string `json:"message"`
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

type Resource struct {
	ID string `json:"id"`
}

type ResourceList struct {
	Data *[]Resource `json:"data,omitempty"`
}

type SendNotificationRequest struct {
	Body         *string        `json:"body,omitempty"`
	Channels     []ChannelsItem `json:"channels"`
	Data         *JsonObject    `json:"data,omitempty"`
	DeviceToken  *string        `json:"device_token,omitempty"`
	Email        *string        `json:"email,omitempty"`
	Locale       *string        `json:"locale,omitempty"`
	Phone        *string        `json:"phone,omitempty"`
	Platform     *string        `json:"platform,omitempty"`
	TemplateData *JsonObject    `json:"template_data,omitempty"`
	TemplateSlug *string        `json:"template_slug,omitempty"`
	Title        *string        `json:"title,omitempty"`
	Type         string         `json:"type"`
	UserID       string         `json:"user_id"`
}

type Template struct {
	Category     *string `json:"category,omitempty"`
	EmailHtml    *string `json:"email_html,omitempty"`
	EmailSubject *string `json:"email_subject,omitempty"`
	ID           string  `json:"id"`
	IsActive     bool    `json:"is_active"`
	Locale       *string `json:"locale,omitempty"`
	Name         string  `json:"name"`
	PushBody     *string `json:"push_body,omitempty"`
	PushTitle    *string `json:"push_title,omitempty"`
	Slug         string  `json:"slug"`
	SmsBody      *string `json:"sms_body,omitempty"`
	Type         *string `json:"type,omitempty"`
	Version      *int64  `json:"version,omitempty"`
}

type TemplateList struct {
	Templates []Template `json:"templates"`
}

type TemplatePreview struct {
	EmailHtml    *string `json:"email_html,omitempty"`
	EmailSubject *string `json:"email_subject,omitempty"`
	Locale       *string `json:"locale,omitempty"`
	PushBody     *string `json:"push_body,omitempty"`
	PushTitle    *string `json:"push_title,omitempty"`
	SmsBody      *string `json:"sms_body,omitempty"`
	Version      *int64  `json:"version,omitempty"`
}

type TemplatePreviewRequest struct {
	Category     *string     `json:"category,omitempty"`
	EmailHtml    *string     `json:"email_html,omitempty"`
	EmailSubject *string     `json:"email_subject,omitempty"`
	Locale       *string     `json:"locale,omitempty"`
	Name         string      `json:"name"`
	PushBody     *string     `json:"push_body,omitempty"`
	PushTitle    *string     `json:"push_title,omitempty"`
	Slug         string      `json:"slug"`
	SmsBody      *string     `json:"sms_body,omitempty"`
	Type         *string     `json:"type,omitempty"`
	Variables    *JsonObject `json:"variables,omitempty"`
	Version      *int64      `json:"version,omitempty"`
}

type TemplateTestRequest struct {
	Channels  []string    `json:"channels"`
	Locale    *string     `json:"locale,omitempty"`
	UserID    string      `json:"user_id"`
	Variables *JsonObject `json:"variables,omitempty"`
}

type TemplateVariables struct {
	Locale    *string     `json:"locale,omitempty"`
	Variables *JsonObject `json:"variables,omitempty"`
}

type TemplateWrite struct {
	Category     *string `json:"category,omitempty"`
	EmailHtml    *string `json:"email_html,omitempty"`
	EmailSubject *string `json:"email_subject,omitempty"`
	Locale       *string `json:"locale,omitempty"`
	Name         string  `json:"name"`
	PushBody     *string `json:"push_body,omitempty"`
	PushTitle    *string `json:"push_title,omitempty"`
	Slug         string  `json:"slug"`
	SmsBody      *string `json:"sms_body,omitempty"`
	Type         *string `json:"type,omitempty"`
	Version      *int64  `json:"version,omitempty"`
}

type ChannelsItem string

const (
	ChannelsItemPush  ChannelsItem = "push"
	ChannelsItemSms   ChannelsItem = "sms"
	ChannelsItemEmail ChannelsItem = "email"
	ChannelsItemInApp ChannelsItem = "in_app"
)

var (
	activateCampaignOperation                = transport.Operation{Name: "activateCampaign", Method: http.MethodPost, Path: "/v1/notifications/campaigns/{id}/activate", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	archiveCampaignOperation                 = transport.Operation{Name: "archiveCampaign", Method: http.MethodPost, Path: "/v1/notifications/campaigns/{id}/archive", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	cancelBroadcastOperation                 = transport.Operation{Name: "cancelBroadcast", Method: http.MethodPost, Path: "/v1/notifications/broadcasts/{id}/cancel", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	completeCampaignOperation                = transport.Operation{Name: "completeCampaign", Method: http.MethodPost, Path: "/v1/notifications/campaigns/{id}/complete", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	createBroadcastOperation                 = transport.Operation{Name: "createBroadcast", Method: http.MethodPost, Path: "/v1/notifications/broadcasts", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{201}}
	createCampaignOperation                  = transport.Operation{Name: "createCampaign", Method: http.MethodPost, Path: "/v1/notifications/campaigns", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{201}}
	createNotificationTemplateOperation      = transport.Operation{Name: "createNotificationTemplate", Method: http.MethodPost, Path: "/v1/notifications/templates", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{201}}
	getBroadcastOperation                    = transport.Operation{Name: "getBroadcast", Method: http.MethodGet, Path: "/v1/notifications/broadcasts/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getBroadcastAnalyticsOperation           = transport.Operation{Name: "getBroadcastAnalytics", Method: http.MethodGet, Path: "/v1/notifications/analytics/broadcasts/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getCampaignOperation                     = transport.Operation{Name: "getCampaign", Method: http.MethodGet, Path: "/v1/notifications/campaigns/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getChannelAnalyticsOperation             = transport.Operation{Name: "getChannelAnalytics", Method: http.MethodGet, Path: "/v1/notifications/analytics/channel/{channel}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getEmailConfigurationOperation           = transport.Operation{Name: "getEmailConfiguration", Method: http.MethodGet, Path: "/v1/notifications/email/config", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getEmailEventOperation                   = transport.Operation{Name: "getEmailEvent", Method: http.MethodGet, Path: "/v1/notifications/email/events/{id}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getNotificationAnalyticsOperation        = transport.Operation{Name: "getNotificationAnalytics", Method: http.MethodGet, Path: "/v1/notifications/analytics", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getNotificationTemplateOperation         = transport.Operation{Name: "getNotificationTemplate", Method: http.MethodGet, Path: "/v1/notifications/templates/{slug}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listBroadcastsOperation                  = transport.Operation{Name: "listBroadcasts", Method: http.MethodGet, Path: "/v1/notifications/broadcasts", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listCampaignExecutionsOperation          = transport.Operation{Name: "listCampaignExecutions", Method: http.MethodGet, Path: "/v1/notifications/campaigns/{id}/executions", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listCampaignsOperation                   = transport.Operation{Name: "listCampaigns", Method: http.MethodGet, Path: "/v1/notifications/campaigns", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listNotificationTemplatesOperation       = transport.Operation{Name: "listNotificationTemplates", Method: http.MethodGet, Path: "/v1/notifications/templates", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	pauseCampaignOperation                   = transport.Operation{Name: "pauseCampaign", Method: http.MethodPost, Path: "/v1/notifications/campaigns/{id}/pause", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	previewNotificationTemplateOperation     = transport.Operation{Name: "previewNotificationTemplate", Method: http.MethodPost, Path: "/v1/notifications/templates/{slug}/preview", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	previewNotificationTemplateBodyOperation = transport.Operation{Name: "previewNotificationTemplateBody", Method: http.MethodPost, Path: "/v1/notifications/templates/preview", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	sendBroadcastOperation                   = transport.Operation{Name: "sendBroadcast", Method: http.MethodPost, Path: "/v1/notifications/broadcasts/{id}/send", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{202}}
	sendEmailOperation                       = transport.Operation{Name: "sendEmail", Method: http.MethodPost, Path: "/v1/notifications/email/send", Idempotency: transport.IdempotencyOptional, SuccessStatuses: []int{202}}
	sendNotificationOperation                = transport.Operation{Name: "sendNotification", Method: http.MethodPost, Path: "/v1/notifications/messages", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{202}}
	testNotificationTemplateOperation        = transport.Operation{Name: "testNotificationTemplate", Method: http.MethodPost, Path: "/v1/notifications/templates/{slug}/test", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	updateCampaignOperation                  = transport.Operation{Name: "updateCampaign", Method: http.MethodPatch, Path: "/v1/notifications/campaigns/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	updateEmailConfigurationOperation        = transport.Operation{Name: "updateEmailConfiguration", Method: http.MethodPut, Path: "/v1/notifications/email/config", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
	updateNotificationTemplateOperation      = transport.Operation{Name: "updateNotificationTemplate", Method: http.MethodPut, Path: "/v1/notifications/templates/{id}", Idempotency: transport.IdempotencyUnsupported, SuccessStatuses: []int{200}}
)

var allOperations = []transport.Operation{
	activateCampaignOperation,
	archiveCampaignOperation,
	cancelBroadcastOperation,
	completeCampaignOperation,
	createBroadcastOperation,
	createCampaignOperation,
	createNotificationTemplateOperation,
	getBroadcastOperation,
	getBroadcastAnalyticsOperation,
	getCampaignOperation,
	getChannelAnalyticsOperation,
	getEmailConfigurationOperation,
	getEmailEventOperation,
	getNotificationAnalyticsOperation,
	getNotificationTemplateOperation,
	listBroadcastsOperation,
	listCampaignExecutionsOperation,
	listCampaignsOperation,
	listNotificationTemplatesOperation,
	pauseCampaignOperation,
	previewNotificationTemplateOperation,
	previewNotificationTemplateBodyOperation,
	sendBroadcastOperation,
	sendEmailOperation,
	sendNotificationOperation,
	testNotificationTemplateOperation,
	updateCampaignOperation,
	updateEmailConfigurationOperation,
	updateNotificationTemplateOperation,
}

func (client *Client) ActivateCampaign(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, activateCampaignOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ArchiveCampaign(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, archiveCampaignOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CancelBroadcast(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, cancelBroadcastOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CompleteCampaign(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, completeCampaignOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) CreateBroadcast(ctx context.Context, request JsonObject, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	var path map[string]string
	var response Resource
	err := client.execute(ctx, createBroadcastOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateCampaign(ctx context.Context, request JsonObject, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	var path map[string]string
	var response Resource
	err := client.execute(ctx, createCampaignOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) CreateNotificationTemplate(ctx context.Context, request TemplateWrite, options ...aether.RequestOption) (*Template, error) {
	var query url.Values
	var path map[string]string
	var response Template
	err := client.execute(ctx, createNotificationTemplateOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) GetBroadcast(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, getBroadcastOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetBroadcastAnalytics(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, getBroadcastAnalyticsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetCampaign(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, getCampaignOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetChannelAnalytics(ctx context.Context, channel ChannelsItem, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"channel": string(channel),
	}
	var response JsonObject
	err := client.execute(ctx, getChannelAnalyticsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetEmailConfiguration(ctx context.Context, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	var path map[string]string
	var response JsonObject
	err := client.execute(ctx, getEmailConfigurationOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetEmailEvent(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, getEmailEventOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetNotificationAnalytics(ctx context.Context, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	var path map[string]string
	var response JsonObject
	err := client.execute(ctx, getNotificationAnalyticsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetNotificationTemplate(ctx context.Context, slug string, options ...aether.RequestOption) (*Template, error) {
	var query url.Values
	path := map[string]string{
		"slug": string(slug),
	}
	var response Template
	err := client.execute(ctx, getNotificationTemplateOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListBroadcasts(ctx context.Context, options ...aether.RequestOption) (*ResourceList, error) {
	var query url.Values
	var path map[string]string
	var response ResourceList
	err := client.execute(ctx, listBroadcastsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListCampaignExecutions(ctx context.Context, id string, options ...aether.RequestOption) (*ResourceList, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response ResourceList
	err := client.execute(ctx, listCampaignExecutionsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListCampaigns(ctx context.Context, options ...aether.RequestOption) (*ResourceList, error) {
	var query url.Values
	var path map[string]string
	var response ResourceList
	err := client.execute(ctx, listCampaignsOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListNotificationTemplates(ctx context.Context, options ...aether.RequestOption) (*TemplateList, error) {
	var query url.Values
	var path map[string]string
	var response TemplateList
	err := client.execute(ctx, listNotificationTemplatesOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) PauseCampaign(ctx context.Context, id string, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, pauseCampaignOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) PreviewNotificationTemplate(ctx context.Context, slug string, request TemplateVariables, options ...aether.RequestOption) (*TemplatePreview, error) {
	var query url.Values
	path := map[string]string{
		"slug": string(slug),
	}
	var response TemplatePreview
	err := client.execute(ctx, previewNotificationTemplateOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) PreviewNotificationTemplateBody(ctx context.Context, request TemplatePreviewRequest, options ...aether.RequestOption) (*TemplatePreview, error) {
	var query url.Values
	var path map[string]string
	var response TemplatePreview
	err := client.execute(ctx, previewNotificationTemplateBodyOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) SendBroadcast(ctx context.Context, id string, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response JsonObject
	err := client.execute(ctx, sendBroadcastOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) SendEmail(ctx context.Context, request JsonObject, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	var path map[string]string
	var response JsonObject
	err := client.execute(ctx, sendEmailOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) SendNotification(ctx context.Context, request SendNotificationRequest, options ...aether.RequestOption) (*QueuedNotification, error) {
	var query url.Values
	var path map[string]string
	var response QueuedNotification
	err := client.execute(ctx, sendNotificationOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) TestNotificationTemplate(ctx context.Context, slug string, request TemplateTestRequest, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	path := map[string]string{
		"slug": string(slug),
	}
	var response JsonObject
	err := client.execute(ctx, testNotificationTemplateOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) UpdateCampaign(ctx context.Context, id string, request JsonObject, options ...aether.RequestOption) (*Resource, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Resource
	err := client.execute(ctx, updateCampaignOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) UpdateEmailConfiguration(ctx context.Context, request JsonObject, options ...aether.RequestOption) (*JsonObject, error) {
	var query url.Values
	var path map[string]string
	var response JsonObject
	err := client.execute(ctx, updateEmailConfigurationOperation, path, query, request, &response, options...)
	return &response, err
}

func (client *Client) UpdateNotificationTemplate(ctx context.Context, id string, request TemplateWrite, options ...aether.RequestOption) (*Template, error) {
	var query url.Values
	path := map[string]string{
		"id": string(id),
	}
	var response Template
	err := client.execute(ctx, updateNotificationTemplateOperation, path, query, request, &response, options...)
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

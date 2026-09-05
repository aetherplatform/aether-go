// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.
// Source: contracts/openapi/v1/events.yaml

package events

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type EventSchema struct {
	Schema  map[string]any `json:"schema"`
	Type    string         `json:"type"`
	Version int64          `json:"version"`
}

type EventType struct {
	CurrentVersion int64   `json:"current_version"`
	Description    string  `json:"description"`
	ID             string  `json:"id"`
	Lifecycle      string  `json:"lifecycle"`
	Name           string  `json:"name"`
	Producer       string  `json:"producer"`
	Versions       []int64 `json:"versions"`
}

type EventTypeDetail struct {
	ConsumerGroups []struct {
		DurableName string `json:"durable_name"`
		Filter      string `json:"filter"`
		MaxDeliver  int64  `json:"max_deliver"`
		Members     int64  `json:"members"`
		Name        string `json:"name"`
		Stream      string `json:"stream"`
	} `json:"consumer_groups"`
	CurrentVersion int64   `json:"current_version"`
	Description    string  `json:"description"`
	ID             string  `json:"id"`
	Lifecycle      string  `json:"lifecycle"`
	Name           string  `json:"name"`
	Producer       string  `json:"producer"`
	Versions       []int64 `json:"versions"`
}

var (
	getEventSchemaOperation = transport.Operation{Name: "getEventSchema", Method: http.MethodGet, Path: "/v1/events/schemas/{type}/{version}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	getEventTypeOperation   = transport.Operation{Name: "getEventType", Method: http.MethodGet, Path: "/v1/events/catalog/{type}", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
	listEventTypesOperation = transport.Operation{Name: "listEventTypes", Method: http.MethodGet, Path: "/v1/events/catalog", Idempotency: transport.IdempotencyNotApplicable, SuccessStatuses: []int{200}}
)

var allOperations = []transport.Operation{
	getEventSchemaOperation,
	getEventTypeOperation,
	listEventTypesOperation,
}

func (client *Client) GetEventSchema(ctx context.Context, typeValue string, version int64, options ...aether.RequestOption) (*struct {
	Data EventSchema `json:"data"`
}, error) {
	var query url.Values
	path := map[string]string{
		"type":    string(typeValue),
		"version": strconv.FormatInt(int64(version), 10),
	}
	var response struct {
		Data EventSchema `json:"data"`
	}
	err := client.execute(ctx, getEventSchemaOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) GetEventType(ctx context.Context, typeValue string, options ...aether.RequestOption) (*struct {
	Data EventTypeDetail `json:"data"`
}, error) {
	var query url.Values
	path := map[string]string{
		"type": string(typeValue),
	}
	var response struct {
		Data EventTypeDetail `json:"data"`
	}
	err := client.execute(ctx, getEventTypeOperation, path, query, nil, &response, options...)
	return &response, err
}

func (client *Client) ListEventTypes(ctx context.Context, options ...aether.RequestOption) (*struct {
	Data []EventType `json:"data"`
}, error) {
	var query url.Values
	var path map[string]string
	var response struct {
		Data []EventType `json:"data"`
	}
	err := client.execute(ctx, listEventTypesOperation, path, query, nil, &response, options...)
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

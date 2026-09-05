package webhooks

import (
	"context"
	"net/url"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/internal/transport"
)

type Client struct {
	transport *transport.Client
}

func NewClient(config aether.Config) (*Client, error) {
	client, err := transport.NewClient(config, "aether-go/"+aether.Version)
	if err != nil {
		return nil, err
	}
	return &Client{transport: client}, nil
}

func (client *Client) execute(
	ctx context.Context,
	operation transport.Operation,
	path map[string]string,
	query url.Values,
	body any,
	result any,
	options ...aether.RequestOption,
) error {
	return client.transport.Execute(ctx, operation, path, query, body, result, options...)
}

# Aether Go SDK Preview

The Aether Go SDK provides typed access to Aether's public APIs. The first
public beta contains the shared Core runtime plus the approved Identity,
Events, Notifications, Storage, and Webhooks public contracts.

The first public preview is available as the immutable `v0.1.0-beta.1` tag:

```bash
go get github.com/aetherplatform/aether-go@v0.1.0-beta.1
```

Go resolves the module through its module proxy or Git repository and records
its version and checksums in your project. No npm account or npm package is
required.

The module supports Go 1.26 and Go 1.27. Use Go 1.26.6 or newer on the 1.26
line, or Go 1.27.1 or newer on the 1.27 line, and keep up with security patches.
Release certification uses Go 1.27.1. Runtime packages use only the Go standard
library; its security fixes reach your application when you rebuild with a
patched Go toolchain.

## Packages

- `identity` — OAuth/OIDC discovery, PKCE authorization URLs, userinfo, and
  server-only confidential-client token, revocation, and introspection calls.
- `events` — three read-only catalog and governed-schema operations.
- `notifications` — 29 customer-facing send, template, broadcast, campaign,
  analytics, and email-configuration operations.
- `storage` — 18 namespace, asset, upload, object, and transfer operations.
- `webhooks` — 19 subscription, delivery, and inbound-endpoint operations plus
  constant-time outbound signature verification.

## Quick Start

```go
package main

import (
	"bytes"
	"context"
	"log"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/clientcredentials"
	"github.com/aetherplatform/aether-go/storage"
)

func main() {
	ctx := context.Background()
	tokens, err := clientcredentials.New(clientcredentials.Config{
		TokenURL:     "https://auth-sandbox.useaether.co/oauth/token",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		Audience:     "aether-storage",
		Capabilities: []string{
			"storage:objects/*:create",
			"storage:objects/*:delete",
			"storage:objects/*:read",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := storage.NewClient(aether.Config{
		BaseURL:      "https://api-sandbox.useaether.co",
		TokenProvider: tokens,
	})
	if err != nil {
		log.Fatal(err)
	}

	payload := []byte("hello from the Aether Go SDK")
	result, err := client.UploadObject(ctx, storage.CreateUploadIntentRequest{
		NamespaceID: "ns_your_namespace",
		LogicalKey:  "examples/hello.txt",
		Filename:    "hello.txt",
		ContentType: "text/plain",
		ExpectedSize: int64(len(payload)),
	}, bytes.NewReader(payload), int64(len(payload)), storage.UploadObjectOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created object %s", result.Object.ID)
}
```

## Credential Boundary

The root and Storage packages contain no client-secret implementation. OAuth
client credentials live in the server-only `clientcredentials` package, which
is excluded from `GOOS=js` builds. Never compile client secrets into browser or
mobile applications.

Identity confidential-client helpers are also excluded from `GOOS=js` builds.
Hosted browser login remains browser-driven; the SDK builds an Authorization
Code with PKCE URL but never accepts a principal ID or substitutes a browser
cookie for an API bearer token.

Requested capabilities are sent through OAuth's `scope` field, but the SDK does
not interpret them as proof of authorization. The access token and Aether
enforcement point remain authoritative.

## Retries

API requests retry bounded network failures, timeouts, HTTP 502/503/504, and
`429 rate_limited` when the operation is contractually safe. The SDK never
retries `429 quota_exceeded` or caller cancellation. A 401 can invalidate one
cached token and trigger one fresh acquisition; the operation is retried only
when its public idempotency contract allows it.

Client-credentials acquisition is independently bounded. Invalid credentials
and non-transient 4xx responses fail immediately.

`MaxRetries: 0` selects the default: two retries for API and Identity requests,
one for client-credentials acquisition. To make exactly one attempt, set
`DisableRetries: true` in `aether.Config`, `identity.Config`, or
`clientcredentials.Config`. This overrides a non-negative `MaxRetries` value;
negative values remain invalid. Configure the token provider separately from
the API client when disabling both layers of retries.

Webhook event publishing retries only when the request contains a nonblank
`IdempotencyKey`. Missing, empty, and whitespace-only keys disable retries.

## Direct Storage Transfers

Presigned upload and download requests intentionally use a separate,
credential-free request path. Aether bearer tokens, request IDs, correlation
IDs, and client secrets are never copied to provider URLs.

Multipart uploads require an `io.ReaderAt` and explicit size, use bounded
workers, preserve provider ETags exactly, and return completed parts in numeric
order.

## Release Status

The `0.x` line is a public preview without an SLA or zero-downtime compatibility
promise. A named release requires passing certification and hosted sandbox
proof for its candidate commit, followed by the protected release workflow.
See `VERSIONING.md` in the exported public repository.

The one-time `bootstrap-release.yml` workflow creates the first immutable beta
tag only when `AETHER_GO_SDK_BOOTSTRAP_RELEASE_ENABLED=true`. After that tag is
verified, disable the bootstrap switch and enable
`AETHER_GO_SDK_RELEASE_ENABLED=true` for reviewed Release Please updates.

Hosted certification requires explicit Identity, Events, Notifications,
Storage, and Webhooks base URLs and audiences, one sandbox confidential client,
and a disposable Storage namespace. The release workflows pass those values
only from the protected `sdk-sandbox` GitHub environment. No hosted credential
is stored in this repository.

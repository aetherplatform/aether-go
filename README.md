# Aether Go SDK Preview

The Aether Go SDK provides typed access to Aether's public APIs. The first
public beta contains the shared Core runtime and the 18 Storage operations
explicitly approved by Aether's public route inventory.

```bash
go get github.com/aetherplatform/aether-go@v0.1.0-beta.1
```

The module supports Go 1.26 and Go 1.27. Release certification uses Go 1.27.1.
Runtime packages use only the Go standard library.

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
		TokenURL:     "https://sandbox.auth.useather.co/oauth/token",
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
		BaseURL:      "https://sandbox.api.useather.co",
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

## Direct Storage Transfers

Presigned upload and download requests intentionally use a separate,
credential-free request path. Aether bearer tokens, request IDs, correlation
IDs, and client secrets are never copied to provider URLs.

Multipart uploads require an `io.ReaderAt` and explicit size, use bounded
workers, preserve provider ETags exactly, and return completed parts in numeric
order.

## Release Status

The `0.x` line is a public preview without an SLA or zero-downtime compatibility
promise. The first Git tag remains blocked until the hosted Storage sandbox
proof succeeds. See `VERSIONING.md` in the exported public repository.

The one-time `bootstrap-release.yml` workflow creates the first immutable beta
tag only when `AETHER_GO_SDK_BOOTSTRAP_RELEASE_ENABLED=true`. After that tag is
verified, disable the bootstrap switch and enable
`AETHER_GO_SDK_RELEASE_ENABLED=true` for reviewed Release Please updates.

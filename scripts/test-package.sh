#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/aether-go-package.XXXXXX")"
trap 'rm -rf "${work_dir}"' EXIT
work_dir="$(cd "${work_dir}" && pwd -P)"

bash "${SDK_ROOT}/scripts/export-public-repository.sh" "${work_dir}/export"
(cd "${work_dir}/export" && bash scripts/check-release-boundary.sh && bash scripts/check-generated.sh)

mkdir -p "${work_dir}/consumer"
cat > "${work_dir}/consumer/go.mod" <<EOF
module example.com/aether-consumer

go 1.26

require github.com/aetherplatform/aether-go v0.0.0
replace github.com/aetherplatform/aether-go => ${work_dir}/export
EOF
cat > "${work_dir}/consumer/main.go" <<'EOF'
package main

import (
	"context"

	"github.com/aetherplatform/aether-go"
	"github.com/aetherplatform/aether-go/clientcredentials"
	"github.com/aetherplatform/aether-go/events"
	"github.com/aetherplatform/aether-go/identity"
	"github.com/aetherplatform/aether-go/notifications"
	"github.com/aetherplatform/aether-go/storage"
	"github.com/aetherplatform/aether-go/webhooks"
)

func main() {
	provider, _ := clientcredentials.New(clientcredentials.Config{
		TokenURL: "https://auth.useather.test/oauth/token", ClientID: "id", ClientSecret: "secret",
		Audience: "aether-storage", Capabilities: []string{"storage:objects/*:read"},
	})
	client, _ := storage.NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: provider})
	_, _ = client.GetObject(context.Background(), storage.ObjectID("obj_example"))
	_, _ = events.NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: provider})
	_, _ = notifications.NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: provider})
	_, _ = webhooks.NewClient(aether.Config{BaseURL: "https://api.useather.test", TokenProvider: provider})
	identityClient, _ := identity.NewClient(identity.Config{BaseURL: "https://auth.useather.test"})
	_, _ = identityClient.AuthorizationURL(identity.AuthorizationRequest{ClientID: "client", RedirectURI: "https://example.com/callback", Scope: "openid", State: "state", CodeChallenge: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
}
EOF

(cd "${work_dir}/export" && go test ./...)
(cd "${work_dir}/consumer" && go build .)

if grep -R -E '/Users/|SERVICE_JWT_PRIVATE_KEY|NATS_URL|platforms/identity|/internal/v1/' \
	"${work_dir}/export" --exclude='check-release-boundary.sh' --exclude='export-public-repository.sh' --exclude='test-package.sh'; then
  echo "private repository material leaked into source-only export" >&2
  exit 1
fi

echo "Source-only Go SDK consumer passed."

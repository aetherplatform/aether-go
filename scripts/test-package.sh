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
	identityClient, _ := identity.NewClient(identity.Config{BaseURL: "https://auth.useather.test", ClientID: "client"})
	proof, _ := identity.GeneratePKCE()
	_, _ = identityClient.StartPasswordless(context.Background(), identity.PasswordlessStartRequest{Identifier: "person@example.test", Channel: identity.ChannelEmail, RedirectURI: "https://example.com/callback", CodeChallenge: proof.Challenge})
	_, _ = identityClient.VerifyPasswordless(context.Background(), identity.PasswordlessVerifyRequest{Transaction: "transaction", ChallengeID: "challenge", Identifier: "person@example.test", Channel: identity.ChannelEmail, Code: "123456", CodeVerifier: proof.Verifier})
	_, _ = identityClient.CompletePasswordless(context.Background(), identity.PasswordlessCompleteRequest{Continuation: "continuation", CodeVerifier: proof.Verifier, Decision: identity.DecisionDeny})
	_, _ = identityClient.ExchangeOAuthToken(context.Background(), identity.TokenRequest{GrantType: identity.GrantAuthorizationCode, Code: "code", RedirectURI: "https://example.com/callback", CodeVerifier: proof.Verifier})
	_ = identityClient.RevokeOAuthToken(context.Background(), identity.TokenHandleRequest{Token: "refresh"})
	_, _ = identity.ValidateOAuthCallback("https://example.com/callback?code=code&state=state", "https://example.com/callback", "state")
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

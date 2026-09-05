#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

grep -q '^module github.com/aetherplatform/aether-go$' "${SDK_ROOT}/go.mod"
grep -q '^go 1\.26$' "${SDK_ROOT}/go.mod"

if grep -Eq '^(replace|exclude|retract) ' "${SDK_ROOT}/go.mod"; then
  echo "public module contains a forbidden go.mod directive" >&2
  exit 1
fi

for platform in events notifications storage webhooks; do
  case "${platform}" in
    events) expected=3 ;;
    notifications) expected=29 ;;
    storage) expected=18 ;;
    webhooks) expected=19 ;;
  esac
  operation_count="$(grep -c '^func (client \*Client) [A-Z].*(ctx context.Context' "${SDK_ROOT}/${platform}/generated.go")"
  if [[ "${operation_count}" != "${expected}" ]]; then
    echo "expected ${expected} generated public ${platform} methods, found ${operation_count}" >&2
    exit 1
  fi
done

for operation in AuthorizationURL GetOpenIDConfiguration GetOAuthJWKS ListOAuthScopes GetUserInfo ExchangeOAuthToken RevokeOAuthToken IntrospectOAuthToken; do
  if ! grep -R -Fq " ${operation}(" "${SDK_ROOT}/identity"; then
    echo "missing public Identity operation ${operation}" >&2
    exit 1
  fi
done

if grep -R -n -E '/oauth/login|/oauth/consent|/internal/v1/' "${SDK_ROOT}/identity"; then
  echo "hosted browser or internal Identity route leaked into the SDK" >&2
  exit 1
fi

if grep -n 'provider-migrations' "${SDK_ROOT}/storage/generated.go"; then
	echo "operator-private Storage route leaked into the generated client" >&2
	exit 1
fi

if grep -R -n -E 'consumer-groups|/v1/notifications/(inbox|devices|preferences)' \
  "${SDK_ROOT}/events/generated.go" "${SDK_ROOT}/notifications/generated.go"; then
  echo "private platform route leaked into a generated client" >&2
  exit 1
fi

if grep -R -E 'SERVICE_JWT_PRIVATE_KEY|NATS_URL|/internal/v1/' \
	"${SDK_ROOT}" --exclude='check-release-boundary.sh' --exclude='export-public-repository.sh' --exclude='test-package.sh'; then
	echo "private platform material leaked into the Go SDK" >&2
	exit 1
fi

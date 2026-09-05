#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

grep -q '^module github.com/aetherplatform/aether-go$' "${SDK_ROOT}/go.mod"
grep -q '^go 1\.26$' "${SDK_ROOT}/go.mod"

if grep -Eq '^(replace|exclude|retract) ' "${SDK_ROOT}/go.mod"; then
  echo "public module contains a forbidden go.mod directive" >&2
  exit 1
fi

operation_count="$(grep -c '^func (client \*Client) [A-Z].*(ctx context.Context' "${SDK_ROOT}/storage/generated.go")"
if [[ "${operation_count}" != "18" ]]; then
  echo "expected 18 generated public Storage methods, found ${operation_count}" >&2
  exit 1
fi

if grep -n 'provider-migrations' "${SDK_ROOT}/storage/generated.go"; then
	echo "operator-private Storage route leaked into the generated client" >&2
	exit 1
fi

if grep -R -E 'SERVICE_JWT_PRIVATE_KEY|NATS_URL|/internal/v1/' \
	"${SDK_ROOT}" --exclude='check-release-boundary.sh' --exclude='export-public-repository.sh' --exclude='test-package.sh'; then
	echo "private platform material leaked into the Go SDK" >&2
	exit 1
fi

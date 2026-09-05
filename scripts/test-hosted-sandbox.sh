#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${SDK_ROOT}"

required=(
  AETHER_SDK_SANDBOX_TOKEN_URL
  AETHER_SDK_SANDBOX_STORAGE_URL
  AETHER_SDK_SANDBOX_CLIENT_ID
  AETHER_SDK_SANDBOX_CLIENT_SECRET
  AETHER_SDK_SANDBOX_NAMESPACE_ID
)
for name in "${required[@]}"; do
  if [[ -z "${!name:-}" ]]; then
    echo "${name} is required" >&2
    exit 64
  fi
done

go run ./cmd/hosted-sandbox-proof

#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AETHER_ROOT="$(cd "${SDK_ROOT}/../.." && pwd)"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/aether-go-sdk-openapi.XXXXXX")"
trap 'rm -rf "${work_dir}"' EXIT

contract="${SDK_ROOT}/contracts/openapi/storage.json"
if [[ ! -f "${contract}" ]]; then
  contract="${work_dir}/storage.json"
  ruby "${AETHER_ROOT}/scripts/openapi_contracts.rb" bundle storage "${contract}" >/dev/null
fi

go run "${SDK_ROOT}/internal/cmd/generate" storage "${contract}" "${SDK_ROOT}/storage/generated.go"

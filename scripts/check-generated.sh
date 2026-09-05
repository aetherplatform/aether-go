#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AETHER_ROOT="$(cd "${SDK_ROOT}/../.." && pwd)"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/aether-go-sdk-openapi.XXXXXX")"
trap 'rm -rf "${work_dir}"' EXIT

for platform in events notifications storage webhooks; do
  temporary="${work_dir}/${platform}.go"
  contract="${SDK_ROOT}/contracts/openapi/${platform}.json"
  if [[ ! -f "${contract}" ]]; then
    contract="${work_dir}/${platform}.json"
    ruby "${AETHER_ROOT}/scripts/openapi_contracts.rb" bundle "${platform}" "${contract}" >/dev/null
  fi
  go run "${SDK_ROOT}/internal/cmd/generate" "${platform}" "${contract}" "${temporary}"
  diff -u "${SDK_ROOT}/${platform}/generated.go" "${temporary}"
done

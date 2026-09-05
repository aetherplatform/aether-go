#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-}"

if [[ -z "${output}" ]]; then
  echo "usage: $0 EMPTY_OUTPUT_DIRECTORY" >&2
  exit 64
fi

mkdir -p "${output}"
if [[ -n "$(find "${output}" -mindepth 1 -maxdepth 1 -print -quit)" ]]; then
  echo "output directory must be empty: ${output}" >&2
  exit 65
fi

rsync -a \
  --exclude coverage.out \
  "${SDK_ROOT}/" "${output}/"

if [[ ! -f "${SDK_ROOT}/contracts/openapi/storage.json" ]]; then
  AETHER_ROOT="$(cd "${SDK_ROOT}/../.." && pwd)"
  mkdir -p "${output}/contracts/openapi" "${output}/contracts/sdk/v1"
  for platform in events identity notifications storage webhooks; do
    ruby "${AETHER_ROOT}/scripts/openapi_contracts.rb" bundle "${platform}" "${output}/contracts/openapi/${platform}.json" >/dev/null
  done
  cp "${AETHER_ROOT}"/contracts/sdk/v1/*.json "${output}/contracts/sdk/v1/"
  mkdir -p "${output}/contracts/webhooks/signatures/v1"
  cp "${AETHER_ROOT}/contracts/webhooks/signatures/v1/vectors.json" "${output}/contracts/webhooks/signatures/v1/vectors.json"
  cp "${AETHER_ROOT}/docs/go-sdk-versioning-and-release-policy.md" "${output}/VERSIONING.md"
fi

if grep -n 'provider-migrations' "${output}/storage/generated.go"; then
	echo "operator-private Storage route leaked into the public Go SDK export" >&2
	exit 66
fi

if grep -R -E '/Users/|SERVICE_JWT_PRIVATE_KEY|NATS_URL|platforms/identity|/internal/v1/' \
	"${output}" --exclude='check-release-boundary.sh' --exclude='export-public-repository.sh' --exclude='test-package.sh'; then
  echo "private repository material leaked into public Go SDK export" >&2
  exit 66
fi

echo "Public Go SDK repository exported to ${output}"

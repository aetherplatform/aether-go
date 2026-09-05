#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${SDK_ROOT}"

bash scripts/check-release-boundary.sh
bash scripts/check-generated.sh
test -z "$(gofmt -l .)"
go mod tidy -diff
go vet ./...
go test ./...
go test -race ./...
bash scripts/test-wasm-boundary.sh
bash scripts/test-local-sandbox.sh
bash scripts/test-package.sh

if command -v staticcheck >/dev/null 2>&1; then
  staticcheck ./...
else
  echo "staticcheck is not installed; CI installs honnef.co/go/tools/cmd/staticcheck@v0.8.1" >&2
fi

if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./...
else
  echo "govulncheck is not installed; CI installs golang.org/x/vuln/cmd/govulncheck@v1.7.0" >&2
fi

#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/aether-go-wasm.XXXXXX")"
trap 'rm -rf "${work_dir}"' EXIT

cd "${SDK_ROOT}"
GOOS=js GOARCH=wasm go test -run '^$' -exec=true . ./events ./identity ./notifications ./storage ./webhooks

mkdir -p "${work_dir}/consumer"
cat > "${work_dir}/consumer/go.mod" <<EOF
module example.com/aether-browser-negative

go 1.26

require github.com/aetherplatform/aether-go v0.0.0
replace github.com/aetherplatform/aether-go => ${SDK_ROOT}
EOF
cat > "${work_dir}/consumer/main.go" <<'EOF'
package main

import _ "github.com/aetherplatform/aether-go/clientcredentials"

func main() {}
EOF

if (cd "${work_dir}/consumer" && GOOS=js GOARCH=wasm go build . >/dev/null 2>&1); then
  echo "clientcredentials unexpectedly compiled for GOOS=js" >&2
  exit 1
fi

mkdir -p "${work_dir}/identity-consumer"
cat > "${work_dir}/identity-consumer/go.mod" <<EOF
module example.com/aether-browser-identity-negative

go 1.26

require github.com/aetherplatform/aether-go v0.0.0
replace github.com/aetherplatform/aether-go => ${SDK_ROOT}
EOF
cat > "${work_dir}/identity-consumer/main.go" <<'EOF'
package main

import "github.com/aetherplatform/aether-go/identity"

func main() {
	_, _ = identity.NewConfidentialClient(identity.ConfidentialConfig{})
}
EOF

if (cd "${work_dir}/identity-consumer" && GOOS=js GOARCH=wasm go build . >/dev/null 2>&1); then
  echo "Identity confidential client unexpectedly compiled for GOOS=js" >&2
  exit 1
fi

#!/usr/bin/env bash

set -euo pipefail

SDK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${SDK_ROOT}"
go test ./identity ./events ./notifications ./webhooks ./storage -count=1

#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Build and run the in-process benchmark binary.
go build -o /dev/null ./cmd/bench/ 2>&1
go run ./cmd/bench/

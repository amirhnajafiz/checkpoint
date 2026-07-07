#!/usr/bin/env bash

# Run static analysis and tests for the project.

set -uo pipefail

# Move to the repo root regardless of where the script was invoked from.
cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Make go-installed tools (golangci-lint, sqlc, ...) reachable.
export PATH="$PATH:$(go env GOPATH)/bin"

status=0
run() {
	echo "==> $*"
	"$@" || status=1
}

# Formatting: gofmt -l prints files that are NOT formatted.
echo "==> gofmt -l ."
fmt_out=$(gofmt -l .)
if [ -n "$fmt_out" ]; then
	echo "these files need 'gofmt -w':"
	echo "$fmt_out"
	status=1
fi

# Vet: catches suspicious constructs the compiler allows.
run go vet ./...

# Lint: golangci-lint aggregates many linters (see .golangci.yml).
if command -v golangci-lint >/dev/null 2>&1; then
	run golangci-lint run ./...
else
	echo "==> golangci-lint not found; skipping"
	echo "    install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
fi

# Tests.
run go test ./...

if [ "$status" -ne 0 ]; then
	echo "checks FAILED"
else
	echo "checks passed"
fi
exit "$status"

#!/usr/bin/env bash

# Run the integration test suite: the `integration`-tagged tests that stand up a
# throwaway PostgreSQL container (testcontainers, needs Docker) and an in-process
# Redis (miniredis), then exercise the store, cache, and HTTP API end to end.
#
#   scripts/integration-test.sh                 # run the whole suite
#   scripts/integration-test.sh -run TestAPI    # forward any `go test` flags
#   scripts/integration-test.sh -v ./internal/http/...
#
# Extra arguments are passed straight through to `go test`. When no package
# pattern is given, every package is tested (./...).

set -euo pipefail

# Run from the repo root regardless of where the script was invoked from.
cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# The integration tests need a reachable Docker daemon for testcontainers.
if ! docker info >/dev/null 2>&1; then
	echo "error: Docker does not appear to be running." >&2
	echo "       Start Docker and retry; these tests spin up a Postgres container." >&2
	exit 1
fi

# Split forwarded args into flags and package patterns so we can default the
# pattern to ./... while still honoring an explicit one.
has_pkg=false
for arg in "$@"; do
	case "$arg" in
	./* | github.com/*) has_pkg=true ;;
	esac
done

echo "==> go test -tags=integration $*"
if [ "$has_pkg" = true ]; then
	go test -tags=integration "$@"
else
	go test -tags=integration "$@" ./...
fi

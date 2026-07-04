#!/usr/bin/env bash
# bring the local mayigoo stack (app + postgres + redis) up or down via
# docker compose.
#   scripts/local-deploy.sh up     # build and start the stack in the background
#   scripts/local-deploy.sh down   # stop and remove the containers

set -euo pipefail

# Run from the repo root so docker compose finds docker-compose.yml and .env.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

action="${1:-}"
shift || true

case "$action" in
  up)
    if [ ! -f .env ]; then
      echo "warning: no .env in repo root; copy example/.env.example to .env and set your OAuth credentials" >&2
    fi
    docker compose up --build -d "$@"
    docker compose ps
    echo "mayigoo is up on http://localhost:5000"
    ;;
  down)
    docker compose down "$@"
    ;;
  *)
    echo "usage: $(basename "$0") {up|down} [extra docker compose args]" >&2
    exit 1
    ;;
esac

# MayIGoo

A small authorization service. Users authenticate via Google OAuth and manage
their own **workspaces**, **roles**, **accounts**, and **role bindings**.
External services call the API to validate their tokens.

## Requirements

These are the tools the project depends on. Versions are what it is currently developed against.

| Tool                                        | Version   | Why it's needed                                               |
| ------------------------------------------- | --------- | ------------------------------------------------------------- |
| [Go](https://go.dev/dl/)                    | `1.26.3`  | Language toolchain (build, `go vet`, `go test`).              |
| [PostgreSQL](https://www.postgresql.org/)   | 13+       | Backing database (accessed via the `pgx` stdlib driver).      |
| [Redis](https://redis.io/)                  | 7+        | Stores the JWT minted for each service account.               |
| [sqlc](https://sqlc.dev/)                   | `v1.31.1` | Generates type-safe Go from the SQL in `internal/db/queries`. |
| [golangci-lint](https://golangci-lint.run/) | `v2.12.2` | Aggregated linting (see `.golangci.yml`).                     |
| [Docker](https://www.docker.com/) + Compose | 24+       | Runs the demo stack (app + Postgres + Redis).                 |
| Google OAuth 2.0 client                     | —         | "Sign in with Google" for users (see below).                  |

> After `go install`, make sure `$(go env GOPATH)/bin` is on your `PATH` so
> `sqlc` and `golangci-lint` are found.

### Google OAuth setup

Login uses Google OAuth, so you need an OAuth 2.0 Client ID from the
[Google Cloud Console](https://console.cloud.google.com/apis/credentials)
(APIs & Services → Credentials → Create credentials → OAuth client ID → Web
application). Then:

1. Add an **Authorized redirect URI** that matches `GOOGLE_REDIRECT_URL`
   **exactly** — for the demo that is:

   ```
   http://localhost:5000/api/users/callback
   ```

   Google compares character-for-character: `http` (not `https`) for localhost,
   no trailing slash, port `5000`, and `localhost` (not `127.0.0.1`) — a
   mismatch yields `Error 400: redirect_uri_mismatch`.

2. Copy `.env.example` to `.env` and fill in `GOOGLE_CLIENT_ID` and
   `GOOGLE_CLIENT_SECRET`.

The app logs its effective `redirect_uri` at startup so you can copy it verbatim
into the Console.

## Run the demo

```bash
cp .env.example .env      # fill in your Google OAuth credentials
docker compose up --build # app on http://localhost:5000, plus Postgres + Redis
```

Open http://localhost:5000, sign in with Google, and manage service accounts.

## Common tasks

**Run all checks** (formatting, vet, lint, tests):

```bash
./scripts/check.sh
```

**Regenerate models after changing SQL** (queries or migrations):

```bash
sqlc generate
```

**Validate the service-account API** end to end (create → list → get token →
validate → rotate-on-update → delete). Sign in first, then copy your user JWT
from the browser (DevTools → Application → Local Storage → `token`):

```bash
USER_TOKEN=<your-jwt> ./scripts/validate.sh
```

A single call, for reference — create an account and read back its service token:

```bash
curl -s -X POST http://localhost:5000/api/accounts \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"demo","kv":{"env":"demo"}}' | jq .
```

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
| [sqlc](https://sqlc.dev/)                   | `v1.31.1` | Generates type-safe Go from the SQL in `internal/db/queries`. |
| [golangci-lint](https://golangci-lint.run/) | `v2.12.2` | Aggregated linting (see `.golangci.yml`).                     |

> After `go install`, make sure `$(go env GOPATH)/bin` is on your `PATH` so
> `sqlc` and `golangci-lint` are found.

## Common tasks

**Run all checks** (formatting, vet, lint, tests):

```bash
./scripts/check.sh
```

**Regenerate models after changing SQL** (queries or migrations):

```bash
sqlc generate
```

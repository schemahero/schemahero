# MEMORY.md - Long-Term Memory

## What SchemaHero Does

- Users declare desired table/view schema as Kubernetes CRDs (or use the kubectl plugin against files)
- The operator reconciles desired vs actual schema and creates `Migration` resources with DDL
- Migrations can be approved manually or applied automatically depending on configuration
- Supports in-cluster and external databases (RDS, Cloud SQL, etc.) via `Database` connection CRDs

## Core Components

- **Manager / operator** — `cmd/manager`, controllers in `pkg/controller/`
- **kubectl-schemahero** — install, plan, apply, approve/reject migrations
- **Plugins** — separate binaries per DB engine under `plugins/`; loaded by manager and CLI
- **CRDs** — `Table`, `Migration`, `View`, `Function`, `Database`, `DatabaseExtension` (v1alpha4)

## Supported databases (plugins)

- postgres, mysql, sqlite, cockroach, cassandra, timescaledb, rqlite

## Testing strategy

- `make test` — Go unit tests for `pkg/` and `cmd/`
- `make test-plugins` — plugin package tests
- `integration/tests/` — end-to-end DDL against real DB containers per engine
- `make envtest` — Kubernetes envtest setup via `hack/envtest.sh`

## Build notes

- `make generate manifests` required after API changes
- `PLUGIN=postgres make install-dev` for local plugin + CLI dev loop
- Nix flake provides `nix develop` dev shell with Go, kubectl, etc.

## Gotchas

- Plugins are separate Go modules — run `go mod tidy` in each when changing deps
- CRD YAML under `config/crds/v1` is generated; don't hand-edit without regenerating
- Integration tests are the source of truth for correct `ALTER` / `CREATE` SQL per engine

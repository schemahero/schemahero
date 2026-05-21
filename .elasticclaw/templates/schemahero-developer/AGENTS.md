# AGENTS.md - Your Workspace

## 🚨 NEVER PUSH TO MAIN — All work through PRs. No exceptions.

This folder is home. Treat it that way.

## Every Session

Before doing anything else:
1. Read `SOUL.md` — this is who you are
2. Read `USER.md` — this is who you're helping
3. Read `memory/` files for recent context

## Memory

You wake up fresh each session. These files are your continuity:
- `memory/YYYY-MM-DD.md` — daily logs of what happened

Capture what matters. Decisions, context, things to remember.

## Safety

- Don't run destructive commands without thinking
- `trash` > `rm` when possible
- When in doubt, ask

## Your Domain

You own SchemaHero. Know it deeply:
- Declarative schema management for Kubernetes and external databases
- The operator, kubectl plugin, and database-specific plugins
- How reconciliation produces and applies DDL (`ALTER TABLE`, etc.)
- Integration tests per database engine

## Key Architecture

### Kubernetes operator (Go)
- `cmd/manager/` — operator entrypoint
- `pkg/controller/` — reconcilers: `table`, `migration`, `database`, `view`, `function`, `databaseextension`
- `pkg/apis/schemas/v1alpha4/` — Table, Migration, View, Function CRDs
- `pkg/apis/databases/v1alpha4/` — Database connection CRDs (postgres, mysql, etc.)
- `pkg/database/` — shared database types and plugin loading
- `config/crds/v1/` — generated CRDs (also copied to `pkg/installer/assets`)
- `config/default/` — kustomize overlays for deployment

### CLI (kubectl plugin)
- `cmd/kubectl-schemahero/` — plugin binary
- `pkg/cli/schemaherokubectlcli/` — `plan`, `apply`, `install`, migration approval, etc.
- `pkg/cli/managercli/` — flags and `run` for the operator

### Database plugins
- `plugins/` — per-engine plugins (postgres, mysql, sqlite, cockroach, cassandra, timescaledb, rqlite)
- Each plugin is a separate Go module with `lib/` for DDL generation and tests
- `make -C plugins <engine>` builds a plugin; `PLUGIN=postgres make install-dev` for local dev

### Integration tests
- `integration/tests/<engine>/` — Docker-based tests with `specs/` (YAML tables), `fixtures.sql`, `expect.sql`
- `integration/tests/fixtures/common.mk` — shared test harness patterns

### Docs and examples
- `examples/` — sample Database/Table YAML for various backends
- Public docs: https://schemahero.io/docs/

## Patterns to Follow

### Fixing operator / reconciliation bugs
1. Find the reconciler in `pkg/controller/<resource>/`
2. Trace from CRD spec → desired schema → plugin call → status updates
3. Add or extend unit tests in `pkg/`; use integration tests when DDL output matters

### Fixing plugin / DDL generation
1. Work in `plugins/<engine>/lib/`
2. Run plugin tests: `make -C plugins test` or engine-specific targets
3. Add or update an integration test under `integration/tests/<engine>/`

### API / CRD changes
1. Edit types in `pkg/apis/`
2. Run `make generate manifests` and commit generated artifacts
3. Consider backward compatibility for existing clusters

### CLI changes
1. Commands live in `pkg/cli/schemaherokubectlcli/`
2. Build with `make bin/kubectl-schemahero`
3. Test with `SCHEMAHERO_PLUGIN_DIR=./plugins/bin ./bin/kubectl-schemahero ...`

# TOOLS.md - Local Notes

## Long-Running Tasks

**Important:** The run timeout is 10 minutes. For tasks that take longer (builds, tests, integration suites), use background processes.

### Pattern for Long Tasks

1. Start in background: `exec with background: true, or yieldMs: 5000`
2. Poll for status: `process action: poll, sessionId: <id>`
3. Check logs: `process action: log, sessionId: <id>`
4. Report progress to the chat periodically

### Examples of Long Tasks
- `make test` / `make test-plugins`
- `integration/tests/**` Docker-based suites
- `make generate manifests`
- Docker image builds (`make local`, `docker build`)

### Don't
- Run blocking commands that take >5 minutes without backgrounding
- Go silent for long periods — send updates

### Do
- Start long commands in background immediately
- Poll every 30-60 seconds
- Update the chat with progress

## Nix

**Nix is preinstalled.** Run builds, tests, and other project commands from a dev shell: `nix develop` at the repo root. Toolchains come from the flake; avoid relying on host `go` outside `nix develop` when possible.

## Build Commands

- `make bin/kubectl-schemahero` — build the kubectl plugin
- `make manager` / `make bin/manager` — build the operator
- `make test` — unit tests (`pkg/`, `cmd/`)
- `make test-plugins` — all plugin module tests
- `make -C plugins <engine>` — build one plugin (e.g. `postgres`)
- `PLUGIN=postgres make install-dev` — dev CLI + plugin in `./bin` and `./plugins/bin`
- `make generate manifests` — regenerate clients, listers, informers, and CRDs
- `make fmt` / `make vet` — format and vet Go code
- `make tidy` — `go mod tidy` across root and plugin modules

## Integration Tests

- Under `integration/tests/<engine>/`
- Typically: `make` in a test directory runs Docker + compares `expect.sql`
- Use when validating DDL or migration behavior for a specific database

## Git and GitHub CLI (`git` / `gh`)

**Use the preinstalled Git credential helper for all authentication.** Do not hunt for tokens, read `gh auth token`, or embed credentials in remote URLs. The environment is already configured so `git` and `gh` obtain credentials through the helper.

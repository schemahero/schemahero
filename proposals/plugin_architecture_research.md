---
date: 2025-08-23T14:47:00-00:00
researcher: Claude
git_commit: 7884409a2ecd46b7582e7a5ac5367d26b2f5f1b4
branch: proposal/plugins
repository: schemahero/schemahero
topic: "Plugin Architecture for Database Engines Research"
tags: [research, codebase, database, plugin, architecture]
status: complete
last_updated: 2025-08-23
last_updated_by: Claude
---

# Research: Plugin Architecture for Database Engines

**Date**: 2025-08-23T14:47:00-00:00  
**Researcher**: Claude  
**Git Commit**: 7884409a2ecd46b7582e7a5ac5367d26b2f5f1b4  
**Branch**: proposal/plugins  
**Repository**: schemahero/schemahero  

## Research Question

Research the current SchemaHero database engine architecture to understand how database drivers are implemented, interfaces defined, and how to transition from in-tree drivers to a plugin-based architecture.

## Summary

SchemaHero currently uses a tightly-coupled in-tree architecture where all database engines (postgres, mysql, cassandra, sqlite, rqlite, timescaledb) are statically linked into the binary. Each database engine implements database-specific connection logic but there's minimal abstraction. The system would benefit significantly from a plugin architecture using hashicorp/go-plugin to enable dynamic loading, easier maintenance, and third-party extensions.

## Detailed Findings

### Current Database Engine Architecture

#### Database Connection Interface
- **Primary Interface**: `pkg/database/interfaces/connection.go` defines `SchemaHeroDatabaseConnection`
- **Interface Methods**:
  - `Close() error`
  - `DatabaseName() string` 
  - `EngineVersion() string`
  - `ListTables() ([]*types.Table, error)`
  - `ListTableForeignKeys(string, string) ([]*types.ForeignKey, error)`
  - `ListTableIndexes(string, string) ([]*types.Index, error)`
  - `GetTablePrimaryKey(string) (*types.KeyConstraint, error)`
  - `GetTableSchema(string) ([]*types.Column, error)`

#### Database Engine Implementations
Each database engine has its own package under `pkg/database/`:

- **Postgres** (`pkg/database/postgres/`): ~15 files including connection.go, deploy.go, create.go, alter.go, tables.go, etc.
- **MySQL** (`pkg/database/mysql/`): ~13 files with similar structure
- **Cassandra** (`pkg/database/cassandra/`): ~8 files 
- **SQLite** (`pkg/database/sqlite/`): ~12 files
- **RQLite** (`pkg/database/rqlite/`): ~12 files  
- **TimescaleDB** (`pkg/database/timescaledb/`): ~4 files (extends postgres)

#### Connection Factories
Each database package provides a `Connect(uri string)` function:
- `postgres.Connect(uri string) (*PostgresConnection, error)`
- `mysql.Connect(uri string) (*MysqlConnection, error)`
- `rqlite.Connect(url string) (*RqliteConnection, error)`
- `sqlite.Connect(dsn string) (*SqliteConnection, error)`
- `cassandra.Connect(hosts []string, username string, password string, keyspace string) (*CassandraConnection, error)`

### Current Database Orchestration

#### Main Database Orchestrator
The `pkg/database/database.go` file contains the main `Database` struct and orchestration logic with large switch statements for each database type:

```go
type Database struct {
    InputDir       string
    OutputDir      string
    Driver         string
    URI            string
    Hosts          []string
    Username       string
    Password       string
    Keyspace       string
    DeploySeedData bool
}
```

#### Database Type Resolution
Database types are determined through CRD connection specifications in `pkg/apis/databases/v1alpha4/database_types.go`:

```go
type DatabaseConnection struct {
    Postgres    *PostgresConnection    `json:"postgres,omitempty"`
    Mysql       *MysqlConnection       `json:"mysql,omitempty"`
    CockroachDB *CockroachDBConnection `json:"cockroachdb,omitempty"`
    Cassandra   *CassandraConnection   `json:"cassandra,omitempty"`
    SQLite      *SqliteConnection      `json:"sqlite,omitempty"`
    RQLite      *RqliteConnection      `json:"rqlite,omitempty"`
    TimescaleDB *PostgresConnection    `json:"timescaledb,omitempty"`
}
```

#### Driver Switch Statements
The current architecture uses large switch statements throughout (`database.go` lines 79-147, 331-347, 364-381, etc.):

```go
if d.Driver == "postgres" {
    return postgres.PlanPostgresTable(d.URI, spec.Name, spec.Schema.Postgres, seedData)
} else if d.Driver == "mysql" {
    return mysql.PlanMysqlTable(d.URI, spec.Name, spec.Schema.Mysql, seedData)
} else if d.Driver == "cockroachdb" {
    return postgres.PlanPostgresTable(d.URI, spec.Name, spec.Schema.CockroachDB, seedData)
// ... continues for each database type
```

### Testing Architecture

#### Integration Testing Structure
Comprehensive integration tests exist under `integration/tests/` organized by database type:
- `integration/tests/postgres/` - 23+ test scenarios
- `integration/tests/mysql/` - 25+ test scenarios  
- `integration/tests/cassandra/` - 7 test scenarios
- `integration/tests/sqlite/` - 23+ test scenarios
- `integration/tests/rqlite/` - 28+ test scenarios
- `integration/tests/timescaledb/` - 13+ test scenarios
- `integration/tests/cockroach/` - 12+ test scenarios

Each test directory contains:
- `Makefile` - Test orchestration
- `fixtures.sql` - Initial database state
- `expect.sql` - Expected migration output
- Database-specific `Dockerfile` files for test environments

## Code References

- `pkg/database/database.go:26-36` - Main Database struct definition
- `pkg/database/interfaces/connection.go:7-19` - SchemaHeroDatabaseConnection interface
- `pkg/database/postgres/connection.go:36-103` - Postgres connection factory
- `pkg/database/mysql/connection.go:30-54` - MySQL connection factory
- `pkg/apis/databases/v1alpha4/database_types.go:24-32` - DatabaseConnection CRD spec
- `pkg/database/database.go:364-381` - Driver switch statement for table planning
- `pkg/database/database.go:456-473` - Driver switch statement for statement execution

## Architecture Insights

### Current Limitations
1. **Tight Coupling**: All database engines are compiled into the main binary
2. **Large Binary Size**: Every deployment includes all database drivers regardless of usage
3. **Maintenance Overhead**: Changes to any database engine require full project rebuild and testing
4. **Extension Difficulty**: Adding new database engines requires core codebase changes
5. **Switch Statement Proliferation**: Driver selection logic is duplicated across many functions

### Existing Extension Points
1. **SchemaHeroDatabaseConnection Interface**: Well-defined interface that could serve as plugin contract
2. **Connection Factory Pattern**: Each driver already implements a Connect function
3. **CRD Connection Spec**: Database types are already abstracted in Kubernetes resources

### Plugin Architecture Opportunities
1. **Interface Compatibility**: The existing `SchemaHeroDatabaseConnection` interface could be extended for plugin use
2. **RPC Communication**: Connection methods are already designed for remote execution (error handling, serializable types)
3. **Dynamic Loading Points**: Switch statements provide clear injection points for plugin resolution
4. **Test Isolation**: Integration tests are already organized by database type, supporting plugin-specific testing

## Historical Context (from proposals/)

No existing proposals found related to plugin architecture or database engine abstraction.

## Open Questions

1. **Backward Compatibility**: How to maintain CRD compatibility while enabling plugin specification?
2. **Plugin Discovery**: How should SchemaHero discover and validate available plugins?
3. **Security Model**: What security controls are needed for plugin execution in Kubernetes environments?
4. **Plugin Lifecycle**: How should plugin installation, updates, and removal be managed?
5. **Performance Impact**: What's the performance overhead of RPC communication vs direct method calls?
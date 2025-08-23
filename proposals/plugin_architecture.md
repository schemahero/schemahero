# Plugin Architecture for Database Engines

**Research Reference**: [plugin_architecture_research.md](./plugin_architecture_research.md)

## TL;DR

Transform SchemaHero from a monolithic architecture with in-tree database engines to a plugin-based system where database drivers are separate, dynamically loaded plugins. This will reduce binary size, simplify maintenance, enable third-party database engine extensions, and speed up development and testing cycles. The change uses hashicorp/go-plugin for RPC communication and maintains full backward compatibility with existing CRDs and user workflows.

## The Problem

SchemaHero currently suffers from several architectural limitations that impede maintenance, extensibility, and development velocity:

1. **Monolithic Binary**: All database engines (postgres, mysql, cassandra, sqlite, rqlite, timescaledb) are compiled into a single binary, resulting in large deployment artifacts regardless of which databases are actually used
2. **Tight Coupling**: Database-specific logic is embedded throughout the codebase via switch statements, making it difficult to modify or extend individual database engines
3. **Maintenance Overhead**: Any change to a single database engine requires rebuilding, testing, and releasing the entire SchemaHero system
4. **Extension Barriers**: Adding support for new database engines (especially proprietary ones) requires modifications to the core SchemaHero codebase
5. **Development Friction**: Testing and developing database-specific features requires running the full integration test suite across all database types
6. **Switch Statement Proliferation**: Driver selection logic is duplicated across multiple functions (`PlanSyncTableSpec`, `ApplySync`, `CreateFixturesSync`, etc.)

This architecture worked well for the initial development phase but now constrains SchemaHero's growth and maintainability as the project matures.

## Prototype / Design

### Architecture Overview

```mermaid
graph TB
    subgraph "SchemaHero Core"
        A[schemahero operator] --> B[Plugin Manager]
        C[kubectl-schemahero] --> B
        B --> D[Plugin Registry]
        B --> E[Plugin Loader]
    end
    
    subgraph "Database Plugins"
        F[postgres-plugin]
        G[mysql-plugin] 
        H[cassandra-plugin]
        I[sqlite-plugin]
        J[custom-plugin]
    end
    
    B --> F
    B --> G
    B --> H
    B --> I
    B --> J
    
    subgraph "Communication"
        K[hashicorp/go-plugin RPC]
    end
    
    B -.-> K
    F -.-> K
    G -.-> K
    H -.-> K
    I -.-> K
    J -.-> K
```

### Plugin Interface

Extend the existing `SchemaHeroDatabaseConnection` interface for plugin communication:

```go
// pkg/database/plugin/interface.go
type DatabasePlugin interface {
    // Metadata
    Name() string
    Version() string
    SupportedEngines() []string
    
    // Connection Management
    Connect(uri string, options map[string]interface{}) (SchemaHeroDatabaseConnection, error)
    Validate(config map[string]interface{}) error
    
    // Plugin Lifecycle
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

### Plugin Registry System (OCI-based with ORAS)

Plugins are distributed as OCI artifacts using ORAS (OCI Registry As Storage), providing a standardized, secure, and widely-supported distribution mechanism.

```go
// pkg/database/plugin/registry.go
type PluginRegistry struct {
    plugins      map[string]*PluginInfo
    loader       *PluginLoader
    orasClient   *oras.Client
    defaultRepo  string // "docker.io/schemahero/plugins"
}

type PluginInfo struct {
    Name        string
    Version     string
    OCIRef      string // e.g., "docker.io/schemahero/plugins/postgres:v1.2.3"
    Digest      string // OCI content digest for verification
    Platform    string // linux/amd64, darwin/arm64, etc.
    Engines     []string
    LocalPath   string // Cached binary location
}

type PluginSource struct {
    Type     string // "oci", "local"
    Registry string // OCI registry URL (for oci type)
    Repo     string // Repository path (for oci type)
    Tag      string // Version tag (for oci type)
    Path     string // File path (for local type)
}
```

**Plugin Resolution Order:**
1. Local file path (if specified) - for development/testing
2. OCI reference (if specified) - for production/custom plugins  
3. Default OCI registry - `docker.io/schemahero/plugins/<name>:<version>`

**Supported Sources:**
- **Local files**: `/path/to/plugin` or `./plugins/postgres` for development
- **Official OCI**: `docker.io/schemahero/plugins/<name>:<version>`
- **Custom OCI**: `<custom-registry>/<repo>/<name>:<version>`
- **Multi-platform**: OCI image index for cross-platform support

### Database Connection Extension

Update CRD to support plugin specification while maintaining backward compatibility:

```yaml
# Optional plugin specification in Database CRD
apiVersion: databases.schemahero.io/v1alpha4
kind: Database
metadata:
  name: my-postgres
spec:
  connection:
    postgres:
      uri: 
        value: "postgres://user:pass@host:5432/db"
    # Optional plugin override (defaults to docker.io/schemahero/plugins/<engine>:latest)
    plugin:
      # Option 1: OCI registry reference
      ociRef: "docker.io/schemahero/plugins/postgres:v1.2.3"
      # Option 2: Local file path (useful for development/testing)
      # localPath: "/var/lib/schemahero/plugins/postgres"
      # Option 3: Custom registry
      # ociRef: "ghcr.io/mycompany/schemahero-plugins/custom-postgres:v2.0.0"
      pullSecret: "registry-credentials" # Optional K8s secret for OCI auth
```

## New Subagents / Commands

### CLI Commands
- `schemahero plugin list` - List available and installed plugins
- `schemahero plugin pull <oci-ref>` - Pull plugin from OCI registry
- `schemahero plugin pull postgres` - Shorthand for official plugin (docker.io/schemahero/plugins/postgres:latest)
- `schemahero plugin pull ghcr.io/company/plugins/custom:v1.0.0` - Pull from custom registry
- `schemahero plugin load <path>` - Load plugin from local file path (for development/testing)
- `schemahero plugin remove <name>` - Remove installed plugin
- `schemahero plugin verify <name>` - Verify plugin (OCI digest or local checksum)

### Subagents
- **plugin-manager** - Handles plugin lifecycle, discovery, and loading
- **plugin-registry** - Manages official and third-party plugin registries
- **plugin-security** - Handles plugin verification, signing, and sandboxing

## Database

**No database schema changes required.** This is purely an architectural refactoring that maintains full CRD compatibility.

## Implementation Plan

### Phase 1: Plugin Interface and Core Infrastructure

**Files to modify:**
- `pkg/database/plugin/` (new directory)
  - `interface.go` - Plugin contract definition
  - `registry.go` - Plugin registry management  
  - `loader.go` - Plugin loading and RPC setup
  - `manager.go` - Plugin lifecycle management
- `pkg/database/database.go` - Add plugin resolution logic
- `cmd/manager/main.go` - Initialize plugin system
- `cmd/kubectl-schemahero/main.go` - Add plugin commands

**Pseudo code for plugin resolution with local support:**

```go
// pkg/database/database.go
func (d *Database) getConnection(ctx context.Context, dbSpec *DatabaseSpec) (interfaces.SchemaHeroDatabaseConnection, error) {
    // Check for explicit plugin configuration
    if dbSpec.Plugin != nil {
        // Try local path first (for development)
        if dbSpec.Plugin.LocalPath != "" {
            return d.pluginManager.LoadLocal(ctx, dbSpec.Plugin.LocalPath)
        }
        
        // Try OCI reference
        if dbSpec.Plugin.OCIRef != "" {
            return d.pluginManager.Pull(ctx, dbSpec.Plugin.OCIRef)
        }
    }
    
    // Try default OCI registry for known engines
    if plugin, err := d.pluginManager.GetDefaultPlugin(ctx, dbSpec.Engine); err == nil {
        return plugin, nil
    }
    
    // Fallback to in-tree drivers for backward compatibility
    return d.getInTreeConnection(dbSpec)
}
```

### Phase 2: Extract In-Tree Drivers to Plugins

**Files to create:**
- `plugins/postgres/` - Postgres plugin implementation
- `plugins/mysql/` - MySQL plugin implementation  
- `plugins/cassandra/` - Cassandra plugin implementation
- `plugins/sqlite/` - SQLite plugin implementation
- `plugins/rqlite/` - RQLite plugin implementation
- `plugins/timescaledb/` - TimescaleDB plugin implementation

**Pseudo code for plugin wrapper:**

```go
// plugins/postgres/main.go
func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: shared.Handshake,
        Plugins: map[string]plugin.Plugin{
            "database": &DatabasePluginImpl{},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}

type DatabasePluginImpl struct{}

func (p *DatabasePluginImpl) Connect(uri string, options map[string]interface{}) (SchemaHeroDatabaseConnection, error) {
    // Wrap existing postgres.Connect functionality
    return postgres.Connect(uri)
}
```

### Phase 3: OCI Registry Integration and Distribution

**Files to create:**
- `pkg/registry/oci_client.go` - ORAS client for OCI operations
- `pkg/registry/plugin_manifest.go` - Plugin metadata and OCI manifest handling
- `scripts/build-and-push-plugins.sh` - Build and push plugins to OCI registry
- `.github/workflows/plugin-release.yml` - Plugin CI/CD with multi-platform builds

**Pseudo code for OCI plugin operations:**

```go
// pkg/registry/oci_client.go
type OCIPluginClient struct {
    store    oras.Store
    registry string
}

func (c *OCIPluginClient) Pull(ctx context.Context, ref string) (*Plugin, error) {
    // Parse OCI reference
    ociRef, err := registry.ParseReference(ref)
    if err != nil {
        // Try default registry for shorthand names
        ref = fmt.Sprintf("docker.io/schemahero/plugins/%s:latest", ref)
        ociRef, err = registry.ParseReference(ref)
    }
    
    // Pull binary artifact using ORAS
    desc, reader, err := oras.Pull(ctx, c.store, ociRef)
    
    // Verify digest
    if !verifyDigest(desc.Digest, reader) {
        return nil, fmt.Errorf("digest verification failed")
    }
    
    // Extract and cache binary
    return extractPlugin(reader, desc.Platform)
}

// LoadLocal loads a plugin from local filesystem (for development/testing)
func (c *OCIPluginClient) LoadLocal(ctx context.Context, path string) (*Plugin, error) {
    // Verify binary exists and is executable
    info, err := os.Stat(path)
    if err != nil {
        return nil, fmt.Errorf("plugin not found at %s: %w", path, err)
    }
    
    if info.Mode()&0111 == 0 {
        return nil, fmt.Errorf("plugin at %s is not executable", path)
    }
    
    // Load plugin metadata
    plugin := &Plugin{
        Path:     path,
        Source:   "local",
        Platform: runtime.GOOS + "/" + runtime.GOARCH,
    }
    
    // Validate plugin interface compatibility
    if err := validatePlugin(plugin); err != nil {
        return nil, fmt.Errorf("invalid plugin: %w", err)
    }
    
    return plugin, nil
}
```

### Phase 4: Security and Verification

**Files to modify:**
- `pkg/database/plugin/security.go` - OCI signature verification using cosign/sigstore
- `pkg/database/plugin/sandbox.go` - Plugin execution sandboxing
- `.github/workflows/plugin-release.yml` - Add cosign signing for all plugin artifacts

**OCI-native security features:**
- Content-addressable storage ensures integrity via digest verification
- Cosign/Sigstore integration for cryptographic signatures
- OCI image signing and SBOM (Software Bill of Materials) attachment
- Registry-level access control and audit logging

**Toggle strategy**: Feature flag `SCHEMAHERO_ENABLE_PLUGINS` to control plugin system activation. Defaults to `false` initially, then `true` after stabilization.

## Testing

### Unit Testing
Follow existing patterns in `pkg/database/*/` directories:
- `pkg/database/plugin/registry_test.go` - Plugin registry tests
- `pkg/database/plugin/loader_test.go` - Plugin loading tests
- `pkg/database/plugin/manager_test.go` - Plugin management tests

### Integration Testing  
Extend existing `integration/tests/` structure:
- `integration/tests/plugins/` - Plugin-specific integration tests
- Modify existing database tests to run in both in-tree and plugin modes
- Add tests for plugin installation, removal, and upgrade scenarios

### End-to-End Testing
- Test complete workflow: plugin installation → database connection → schema migration
- Test plugin fallback scenarios when plugins are unavailable
- Test mixed environments with some in-tree and some plugin-based connections

### Plugin-Specific Testing
Each plugin will have its own test suite:
- `plugins/postgres/integration_test.go`
- `plugins/mysql/integration_test.go`
- etc.

## Backward Compatibility

### API Compatibility
- **CRDs remain unchanged** - Existing Database, Table, and Migration resources work without modification
- **CLI compatibility** - Existing `kubectl schemahero` commands function identically
- **Configuration compatibility** - Database connection specifications remain the same

### Migration Path
1. **Phase 1**: Plugin system runs alongside in-tree drivers (no user impact)
2. **Phase 2**: Users can optionally enable plugin mode via feature flag
3. **Phase 3**: Plugin mode becomes default, in-tree drivers deprecated
4. **Phase 4**: In-tree drivers removed in major version release (1+ year timeline)

### Fallback Mechanism
```go
// Graceful degradation when plugins unavailable
if pluginConnection, err := pluginManager.GetConnection(ctx, dbSpec); err != nil {
    logger.Warn("Plugin connection failed, falling back to in-tree driver")
    return d.getInTreeConnection(dbSpec)
}
```

## Migrations

### Operational Steps

1. **Deploy updated SchemaHero operator** with plugin support (feature flag disabled)
2. **Install official plugins** using `schemahero plugin install postgres mysql ...`
3. **Enable plugin mode** by setting `SCHEMAHERO_ENABLE_PLUGINS=true`
4. **Verify all connections work** through existing Database resources
5. **Gradually remove in-tree drivers** in subsequent releases

### Plugin Installation Automation

For hardened environments that can't pull plugins at runtime:

```yaml
# Plugin installation via init container using ORAS
spec:
  template:
    spec:
      initContainers:
      - name: pull-plugins
        image: schemahero/schemahero:latest
        command: 
        - sh
        - -c
        - |
          schemahero plugin pull docker.io/schemahero/plugins/postgres:v1.2.3
          schemahero plugin pull docker.io/schemahero/plugins/mysql:v1.2.3
          # Or pull from private registry
          # schemahero plugin pull registry.company.com/db/plugins/postgres:stable
        volumeMounts:
        - name: plugin-cache
          mountPath: /var/lib/schemahero/plugins
        env:
        - name: REGISTRY_AUTH_FILE
          value: /var/run/secrets/registry/.dockerconfigjson
      volumes:
      - name: plugin-cache
        emptyDir: {}
      # For private registries
      - name: registry-auth
        secret:
          secretName: registry-credentials
```

**Air-gapped environments:**
- Pre-pull plugins to private OCI registry
- Configure SchemaHero to use private registry as default
- Use `imagePullSecrets` for authentication

**No special deployment handling required** - the change is backward compatible and can be rolled out incrementally.

## Trade-offs

### Why This Approach Over Alternatives

**Optimizing for**: Maintainability, extensibility, binary size reduction, development velocity

**Trade-offs accepted**:
- **Performance**: Small RPC overhead vs direct method calls (acceptable for schema operations)
- **Complexity**: Plugin management complexity vs monolithic simplicity
- **Deployment**: Multi-binary deployment vs single binary (mitigated by automation)

### Benefits vs Costs
- ✅ **Reduced maintenance** - Database engines can be updated independently
- ✅ **Faster development** - Test only relevant database engines during development  
- ✅ **Third-party extensibility** - Proprietary database support without core changes
- ✅ **Smaller deployments** - Only deploy needed database drivers
- ❌ **Increased complexity** - Plugin management and RPC communication overhead
- ❌ **More moving parts** - Plugin compatibility, versioning, and distribution

## Alternative Solutions Considered

### 1. Go Build Tags
**Considered**: Using build tags to conditionally compile database engines
**Rejected**: Still requires core codebase changes for new engines, doesn't enable runtime plugin loading

### 2. WebAssembly (WASM) Plugins
**Considered**: WASM-based plugin system for maximum portability
**Rejected**: Limited ecosystem, performance concerns, complexity of database driver porting to WASM

### 3. Container-Based Drivers
**Considered**: Each database engine as separate microservice
**Rejected**: Excessive network overhead for schema operations, operational complexity

### 4. Lua/JavaScript Scripting
**Considered**: Embedded scripting language for database extensions
**Rejected**: Limited to simple extensions, can't leverage existing Go database libraries

## Research

### Prior Art in Codebase
- **Interface-based design**: `pkg/database/interfaces/connection.go` already provides abstraction layer
- **Connection factories**: Each database package already implements `Connect()` pattern
- **Switch statement patterns**: Clear plugin injection points identified throughout codebase

### External References
- **hashicorp/go-plugin**: Battle-tested plugin system used by Terraform, Vault, Packer
- **gRPC communication**: Proven for high-performance RPC in distributed systems
- **Kubernetes operator patterns**: Plugin installation via init containers is established pattern

### Plugin Architecture Examples
- **Terraform providers**: Similar database plugin pattern with provider registry
- **kubectl plugins**: Kubernetes plugin discovery and execution model
- **Docker plugins**: Volume and network plugin architecture

## Checkpoints (PR Plan)

### PR 1: Plugin Infrastructure Foundation
- Plugin interface definitions
- Registry and loader implementation  
- Plugin manager with fallback logic
- Unit tests for core plugin functionality
- Feature flag implementation

### PR 2: hashicorp/go-plugin Integration
- RPC protocol definition using go-plugin
- Plugin handshake and communication setup
- Integration tests for plugin communication
- Error handling and connection management

### PR 3: First Database Plugin (Postgres)
- Extract postgres package into standalone plugin
- Plugin wrapper implementation
- Integration tests for plugin mode vs in-tree mode
- Plugin build and distribution scripts

### PR 4: Remaining Official Plugins
- Extract mysql, cassandra, sqlite, rqlite, timescaledb
- Plugin registry with official plugin definitions
- End-to-end testing across all plugin types
- Documentation for plugin development

### PR 5: CLI Plugin Management
- `schemahero plugin` command implementation
- Plugin installation and removal workflows
- Plugin verification and security checks
- User documentation and migration guide

### PR 6: Security and Production Readiness
- Plugin signature verification
- Sandboxing and security controls
- Performance optimizations
- Production deployment documentation

This checkpoint approach allows for incremental delivery with natural review boundaries while maintaining system stability throughout the migration.
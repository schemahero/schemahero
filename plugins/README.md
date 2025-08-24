# SchemaHero Database Plugins

This directory contains database plugins for SchemaHero that implement the plugin architecture.

## Architecture

SchemaHero uses HashiCorp's go-plugin framework to enable database drivers to be loaded as separate binaries. This provides:

- **Reduced binary size** - Only load the database drivers you need
- **Independent updates** - Update database drivers without rebuilding SchemaHero
- **Third-party extensions** - Add support for proprietary databases without modifying core
- **Faster development** - Test and develop database drivers independently

## Available Plugins

| Plugin | Database | Status | Supported Engines |
|--------|----------|--------|-------------------|
| postgres | PostgreSQL | ✅ Ready | postgres, postgresql, cockroachdb |
| mysql | MySQL/MariaDB | 🚧 Coming Soon | mysql, mariadb |
| cassandra | Cassandra | 🚧 Coming Soon | cassandra |
| sqlite | SQLite | 🚧 Coming Soon | sqlite |
| rqlite | RQLite | 🚧 Coming Soon | rqlite |
| timescaledb | TimescaleDB | 🚧 Coming Soon | timescaledb |

## Building Plugins

### From Root Directory
```bash
# Build all plugins
make plugins

# Test all plugins
make test-plugins
```

### From Plugins Directory
```bash
cd plugins/

# Build all plugins
make all

# Build specific plugin
make postgres
make mysql

# Clean build artifacts
make clean

# Test all plugins
make test

# Test specific plugin
make test-postgres
make test-mysql
```

## Using Plugins

### Manual Plugin Loading

Place plugin binaries in `/var/lib/schemahero/plugins/` or specify a custom path:

```yaml
apiVersion: databases.schemahero.io/v1alpha4
kind: Database
metadata:
  name: my-postgres
spec:
  connection:
    postgres:
      uri: 
        value: "postgres://user:pass@host:5432/db"
    plugin:
      localPath: "/path/to/schemahero-postgres"
```

### Automatic Plugin Discovery

SchemaHero automatically attempts to load plugins from:
1. `/var/lib/schemahero/plugins/` - Default plugin directory
2. Custom paths specified in Database CRDs
3. OCI registries (future feature)

## Plugin Development

### Creating a New Plugin

1. Create a new directory under `plugins/`
2. Implement the `DatabasePlugin` interface from `pkg/database/plugin/interface.go`
3. Create a main.go that serves the plugin using go-plugin
4. Add build target to the Makefile

### Plugin Interface

```go
type DatabasePlugin interface {
    Name() string
    Version() string
    SupportedEngines() []string
    Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error)
    Validate(config map[string]interface{}) error
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

### Example Plugin Structure

```
plugins/
├── mydb/
│   ├── main.go       # Plugin entry point
│   ├── plugin.go     # DatabasePlugin implementation
│   ├── plugin_test.go # Integration tests
│   └── build.sh      # Build script
```

### Testing Plugins

Run plugin tests:
```bash
cd plugins/postgres
go test -v ./...
```

Run integration tests:
```bash
make test
```

## Deployment

### Kubernetes

Deploy plugins as init containers to pre-populate the plugin directory:

```yaml
spec:
  template:
    spec:
      initContainers:
      - name: install-plugins
        image: schemahero/schemahero:latest
        command: 
        - sh
        - -c
        - |
          cp /plugins/* /var/lib/schemahero/plugins/
        volumeMounts:
        - name: plugin-dir
          mountPath: /var/lib/schemahero/plugins
      containers:
      - name: schemahero
        volumeMounts:
        - name: plugin-dir
          mountPath: /var/lib/schemahero/plugins
      volumes:
      - name: plugin-dir
        emptyDir: {}
```

### Docker

Mount plugin directory:
```bash
docker run -v /path/to/plugins:/var/lib/schemahero/plugins schemahero/schemahero
```

## Plugin Security

- Plugins run as separate processes with limited permissions
- Communication happens via RPC with authentication handshake
- Plugins cannot access SchemaHero's memory space
- Resource limits can be applied to plugin processes

## Troubleshooting

### Plugin Not Loading

1. Check plugin binary is executable: `chmod +x plugin-binary`
2. Verify plugin implements correct interface
3. Check logs for handshake errors
4. Ensure plugin version compatibility

### Connection Failures

1. Verify database credentials and connectivity
2. Check plugin supports the database engine
3. Review plugin logs for detailed errors

## Contributing

Contributions are welcome! Please:

1. Follow the existing plugin structure
2. Include comprehensive tests
3. Update documentation
4. Test with actual database instances

## License

Plugins inherit SchemaHero's Apache 2.0 license.
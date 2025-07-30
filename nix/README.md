# SchemaHero Nix Package

This directory contains a Nix package for SchemaHero CLI that can be used locally and proposed to nixpkgs.

## Quick Start

### Using the Flake (Recommended)

```bash
# Run SchemaHero directly
nix run github:schemahero/schemahero#schemahero -- --help

# Install in your environment
nix profile install github:schemahero/schemahero#schemahero

# Use in a devShell
nix develop github:schemahero/schemahero
```

### Using from this repository

```bash
# Build the package
nix build .#schemahero

# Run the CLI
nix run .#schemahero -- --help

# Run as kubectl plugin
nix run .#kubectl-schemahero -- --help

# Enter development shell with SchemaHero available
nix develop
```

## Package Features

- ✅ **Both binaries**: Provides both `schemahero` and `kubectl-schemahero` commands
- ✅ **Static binary**: Built with CGO_ENABLED=0 for maximum compatibility
- ✅ **Version injection**: Proper version, git SHA, and build time in binary
- ✅ **Flake support**: Modern Nix flakes for easy consumption
- ✅ **Development shell**: Complete dev environment with Go, make, kubectl
- ✅ **CI tested**: Automated testing on every change and release

## Maintaining the Package

### Updating to a New Version

1. **Automatic (on release)**:
   The GitHub Actions workflow will create a PR with version update when a new release is published.

2. **Manual update**:
   ```bash
   # Update to latest tag
   ./nix/update-hashes.sh
   
   # Update to specific version
   ./nix/update-hashes.sh 0.21.0
   ```

3. **Test the update**:
   ```bash
   nix build .#schemahero
   nix run .#schemahero -- version
   nix flake check
   ```

### Hash Updates

The script `./nix/update-hashes.sh` automatically:
- Updates the version number
- Fetches and updates the source hash
- Builds the package to get the vendor hash
- Updates the vendor hash
- Tests the final package

## For nixpkgs Submission

To submit this package to nixpkgs:

1. **Copy the package file**:
   ```bash
   cp nix/schemahero.nix /path/to/nixpkgs/pkgs/applications/misc/schemahero/default.nix
   ```

2. **Update all-packages.nix**:
   ```nix
   schemahero = callPackage ../applications/misc/schemahero { };
   ```

3. **Add yourself as maintainer** in the package file:
   ```nix
   maintainers = with maintainers; [ your-github-username ];
   ```

4. **Test in nixpkgs**:
   ```bash
   nix-build -A schemahero
   ```

5. **Create PR** with title: `schemahero: init at 0.21.0`

## Package Structure

```
nix/
├── README.md           # This documentation
├── schemahero.nix      # Main package definition (nixpkgs-ready)
└── update-hashes.sh    # Helper script for maintenance

flake.nix               # Nix flake for modern usage
.github/workflows/nix.yml # CI for testing package
```

## Development

The package includes a development shell with all necessary tools:

```bash
nix develop
# Now you have: go, make, kubectl, and schemahero available
```

This makes it easy to:
- Test changes to the SchemaHero source
- Compare nix-built vs make-built binaries
- Develop and test the Nix package itself

## Testing

The flake includes comprehensive checks:

```bash
# Run all checks
nix flake check

# Individual checks
nix build .#checks.x86_64-linux.build         # Build test
nix build .#checks.x86_64-linux.schemahero-test # Functionality test
```

## License

Same as SchemaHero: Apache License 2.0 

#!/bin/bash

# Script to build and push SchemaHero plugins as OCI artifacts to DockerHub
# Usage: ./scripts/push-plugins.sh [plugin-name] [version]
# Example: ./scripts/push-plugins.sh postgres 0.0.1
# Example: ./scripts/push-plugins.sh all 0.0.1

set -e

REGISTRY="${REGISTRY:-docker.io}"
REGISTRY_NAMESPACE="${REGISTRY_NAMESPACE:-schemahero}"
PLUGIN_VERSION="${2:-0.0.1}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_requirements() {
    local missing_tools=()
    
    if ! command -v oras &> /dev/null; then
        missing_tools+=("oras")
    fi
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_info "Please install missing tools:"
        
        if [[ " ${missing_tools[*]} " =~ " oras " ]]; then
            print_info "  Install ORAS: https://oras.land/docs/installation"
            print_info "  Or run: brew install oras (on macOS)"
        fi
        
        if [[ " ${missing_tools[*]} " =~ " go " ]]; then
            print_info "  Install Go: https://golang.org/doc/install"
        fi
        
        exit 1
    fi
}

# Function to build a plugin for multiple platforms
build_plugin() {
    local plugin_name=$1
    local plugin_binary="schemahero-${plugin_name}"
    
    print_info "Building plugin: ${plugin_name}"
    
    # Create dist directory
    mkdir -p dist
    
    # Build for each platform
    IFS=',' read -ra PLATFORM_ARRAY <<< "$PLATFORMS"
    for platform in "${PLATFORM_ARRAY[@]}"; do
        OS=$(echo $platform | cut -d'/' -f1)
        ARCH=$(echo $platform | cut -d'/' -f2)
        
        print_info "  Building for ${platform}..."
        
        # Build the plugin
        (cd plugins/${plugin_name} && \
         CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -o ../../dist/${plugin_binary}-${OS}-${ARCH} .)
        
        # Create tarball
        tar -czf dist/${plugin_binary}-${OS}-${ARCH}.tar.gz -C dist ${plugin_binary}-${OS}-${ARCH}
        
        # Create checksum
        (cd dist && sha256sum ${plugin_binary}-${OS}-${ARCH}.tar.gz > ${plugin_binary}-${OS}-${ARCH}.tar.gz.sha256)
        
        print_info "    ✓ Built ${plugin_binary}-${OS}-${ARCH}"
    done
    
    # Create manifest
    create_manifest $plugin_name
}

# Function to create a manifest for the plugin
create_manifest() {
    local plugin_name=$1
    local plugin_binary="schemahero-${plugin_name}"
    
    cat > dist/manifest.json <<EOF
{
  "plugin": "${plugin_name}",
  "version": "${PLUGIN_VERSION}",
  "platforms": "${PLATFORMS}",
  "artifacts": [
EOF
    
    IFS=',' read -ra PLATFORM_ARRAY <<< "$PLATFORMS"
    first=true
    for platform in "${PLATFORM_ARRAY[@]}"; do
        OS=$(echo $platform | cut -d'/' -f1)
        ARCH=$(echo $platform | cut -d'/' -f2)
        
        if [ "$first" = false ]; then
            echo "," >> dist/manifest.json
        fi
        echo -n "    {\"platform\": \"$platform\", \"file\": \"${plugin_binary}-${OS}-${ARCH}.tar.gz\"}" >> dist/manifest.json
        first=false
    done
    
    cat >> dist/manifest.json <<EOF

  ]
}
EOF
}

# Function to push plugin to registry as OCI artifact
push_plugin() {
    local plugin_name=$1
    local plugin_binary="schemahero-${plugin_name}"
    local oci_repo="${REGISTRY}/${REGISTRY_NAMESPACE}/plugins/${plugin_name}"
    
    print_info "Pushing plugin ${plugin_name} to ${oci_repo}"
    
    # Check if logged in to registry
    if ! oras manifest fetch "${oci_repo}:latest" &> /dev/null; then
        print_warn "Not logged in to ${REGISTRY} or repository doesn't exist yet"
        print_info "Please run: docker login ${REGISTRY}"
        read -p "Press enter to continue after logging in, or Ctrl-C to cancel..."
    fi
    
    # Push each platform artifact
    IFS=',' read -ra PLATFORM_ARRAY <<< "$PLATFORMS"
    for platform in "${PLATFORM_ARRAY[@]}"; do
        OS=$(echo $platform | cut -d'/' -f1)
        ARCH=$(echo $platform | cut -d'/' -f2)
        
        print_info "  Pushing ${platform} artifact..."
        
        oras push "${oci_repo}:${PLUGIN_VERSION}-${OS}-${ARCH}" \
          --artifact-type application/vnd.schemahero.plugin.v1+tar \
          dist/${plugin_binary}-${OS}-${ARCH}.tar.gz:application/gzip \
          dist/${plugin_binary}-${OS}-${ARCH}.tar.gz.sha256:text/plain \
          dist/manifest.json:application/json \
          --annotation "org.opencontainers.image.title=${plugin_binary}" \
          --annotation "org.opencontainers.image.version=${PLUGIN_VERSION}" \
          --annotation "org.opencontainers.image.description=SchemaHero ${plugin_name} database plugin" \
          --annotation "org.opencontainers.image.source=https://github.com/schemahero/schemahero" \
          --annotation "org.opencontainers.image.platform=${platform}"
        
        print_info "    ✓ Pushed ${platform}"
    done
    
    # Create multi-platform tag
    print_info "  Creating multi-platform tag ${PLUGIN_VERSION}..."
    oras tag "${oci_repo}:${PLUGIN_VERSION}-linux-amd64" "${PLUGIN_VERSION}"
    
    # Tag as latest
    print_info "  Tagging as latest..."
    oras tag "${oci_repo}:${PLUGIN_VERSION}" latest
    
    print_info "✓ Successfully pushed ${plugin_name} v${PLUGIN_VERSION}"
}

# Main script logic
main() {
    local plugin_name=$1
    
    if [ -z "$plugin_name" ]; then
        print_error "Usage: $0 <plugin-name|all> [version]"
        print_info "Available plugins: postgres, mysql, timescaledb, sqlite, rqlite, cassandra"
        exit 1
    fi
    
    check_requirements
    
    # Get list of plugins to build
    if [ "$plugin_name" = "all" ]; then
        plugins=("postgres" "mysql" "timescaledb" "sqlite" "rqlite" "cassandra")
    else
        plugins=("$plugin_name")
    fi
    
    # Clean dist directory
    rm -rf dist
    mkdir -p dist
    
    # Build and push each plugin
    for plugin in "${plugins[@]}"; do
        if [ ! -d "plugins/${plugin}" ]; then
            print_error "Plugin directory not found: plugins/${plugin}"
            exit 1
        fi
        
        build_plugin "$plugin"
        
        # Ask before pushing
        read -p "Push ${plugin} to registry? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            push_plugin "$plugin"
        else
            print_info "Skipping push for ${plugin}"
        fi
        
        # Clean up binaries but keep tarballs for inspection
        rm -f dist/schemahero-${plugin}-*[^.tar.gz]
    done
    
    print_info "✓ All done!"
    print_info "Artifacts are in dist/ directory"
}

# Run main function
main "$@"
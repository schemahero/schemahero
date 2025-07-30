#!/usr/bin/env bash

set -euo pipefail

# Helper script to update hashes for SchemaHero Nix package
# Usage: ./nix/update-hashes.sh [version]

VERSION=${1:-$(git describe --tags --abbrev=0 | sed 's/^v//')}

echo "Updating SchemaHero Nix package to version: $VERSION"

NIX_FILE="nix/schemahero.nix"

# Update version
sed -i.bak "s/version = \".*\";/version = \"$VERSION\";/" "$NIX_FILE"

echo "Getting source hash..."
# Get source hash
echo "Fetching source..."
nix-prefetch-url --unpack "https://github.com/schemahero/schemahero/archive/v${VERSION}.tar.gz" > /tmp/source_hash.txt
SOURCE_HASH_RAW=$(cat /tmp/source_hash.txt)
SOURCE_HASH=$(nix hash to-sri --type sha256 "$SOURCE_HASH_RAW")
echo "Source hash: $SOURCE_HASH"

# Update source hash
sed -i.bak "s/hash = \".*\";/hash = \"$SOURCE_HASH\";/" "$NIX_FILE"

echo "Getting vendor hash (this may take a moment)..."
# To get vendor hash, we need to try building and capture the expected hash from error
if ! VENDOR_HASH_OUTPUT=$(nix build --no-link --print-build-logs .#schemahero 2>&1); then
    if echo "$VENDOR_HASH_OUTPUT" | grep -q "got:.*sha256-"; then
        VENDOR_HASH=$(echo "$VENDOR_HASH_OUTPUT" | grep -o "got:.*sha256-[A-Za-z0-9+/=]*" | cut -d: -f2 | xargs)
        echo "Vendor hash: $VENDOR_HASH"

        # Update vendor hash
        sed -i.bak "s|vendorHash = \"sha256-.*\";|vendorHash = \"$VENDOR_HASH\";|" "$NIX_FILE"

        echo "Attempting build with updated vendor hash..."
        if nix build --no-link .#schemahero; then
            echo "✅ Package builds successfully!"
        else
            echo "❌ Package still fails to build"
            exit 1
        fi
    else
        echo "❌ Could not extract vendor hash from build output"
        echo "Build output:"
        echo "$VENDOR_HASH_OUTPUT"
        exit 1
    fi
else
    echo "✅ Package builds successfully (no vendor hash update needed)!"
fi

# Clean up backup files
rm -f "$NIX_FILE.bak"

echo "Hash update complete! Testing package..."
nix run .#schemahero -- version

echo "✅ All done! Package is ready."
echo ""
echo "To test the package:"
echo "  nix build .#schemahero"
echo "  nix run .#schemahero -- --help"
echo ""
echo "For nixpkgs submission, copy the contents of nix/schemahero.nix"

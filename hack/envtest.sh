#!/usr/bin/env bash

set -euo pipefail

# Install envtest binaries (etcd, kube-apiserver, kubectl) using setup-envtest.
# This replaces the old approach of downloading from Google Cloud Storage,
# which is no longer accessible.

ENVTEST_K8S_VERSION=${ENVTEST_K8S_VERSION:-1.28.x}

if [[ -z "${TMPDIR:-}" ]]; then
    TMPDIR=/tmp
fi

DEST="${TMPDIR}/kubebuilder"

# Install setup-envtest if not present
if ! command -v setup-envtest &> /dev/null; then
    echo "Installing setup-envtest..."
    go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
fi

# Download envtest binaries
echo "Setting up envtest binaries for Kubernetes ${ENVTEST_K8S_VERSION}..."
ENVTEST_ASSETS=$(setup-envtest use "${ENVTEST_K8S_VERSION}" -p path)

# Link to the expected location for the test suite
rm -rf "${DEST}"
mkdir -p "${DEST}"
ln -sf "${ENVTEST_ASSETS}" "${DEST}/bin"

echo "Envtest binaries available at: ${DEST}/bin"

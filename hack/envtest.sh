#!/usr/bin/env bash

# Install etcd and kube-api
#
# Using the same version that controller-runtime
# uses, currently the way envtest invokes the kube-apiserver
# uses flags that have been deprecated in k8s 1.20+, tried
# working around it but it was a hassle and likely rundundant
# work presuming controller-runtime will fix this eventually
version=1.19.2
download_url=https://storage.googleapis.com/kubebuilder-tools
if [[ -z "${TMPDIR}" ]]; then
    TMPDIR=/tmp
fi
if [[ "$OSTYPE" == "darwin"* ]]; then
    rm -f ${TMPDIR}/kubebuilder-tools-${version}-darwin-amd64.tar.gz
    rm -rf ${TMPDIR}/kubebuilder && mkdir -p ${TMPDIR}/kubebuilder

    curl -L ${download_url}/kubebuilder-tools-${version}-darwin-amd64.tar.gz -o ${TMPDIR}/kubebuilder-tools-${version}-darwin-amd64.tar.gz
    tar -xzvf ${TMPDIR}/kubebuilder-tools-${version}-darwin-amd64.tar.gz -C $TMPDIR
else
    rm -f ${TMPDIR}/kubebuilder-tools-${version}-linux-amd64.tar.gz
    rm -rf ${TMPDIR}/kubebuilder && mkdir -p ${TMPDIR}/kubebuilder

    curl -L ${download_url}/kubebuilder-tools-${version}-linux-amd64.tar.gz -o ${TMPDIR}/kubebuilder-tools-${version}-linux-amd64.tar.gz
    tar -xzvf ${TMPDIR}/kubebuilder-tools-${version}-linux-amd64.tar.gz -C $TMPDIR
fi

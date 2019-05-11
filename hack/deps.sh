#!/bin/bash

# Install Kubebuilder
version=1.0.8 # latest stable version
arch=amd64
curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${version}/kubebuilder_${version}_linux_${arch}.tar.gz"
tar -zxvf kubebuilder_${version}_linux_${arch}.tar.gz
mv kubebuilder_${version}_linux_${arch} kubebuilder && sudo mv kubebuilder /usr/local/

# Install kind
GO111MODULE="on" go get -u sigs.k8s.io/kind@65abdce

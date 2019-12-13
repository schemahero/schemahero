# Contributing to SchemaHero

This will eventually be the guide describing how to contribute to SchemaHero and how to set up a dev environment.

The recommended dev workflow is to run a lightweight local Kubernetes cluster such as [microk8s](https://microk8s.io/), with `kubectl` configured with that cluster as the default context, and then to run `make install` to build and deploy your local changes to that cluster.

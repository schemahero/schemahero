# SchemaHero

[![Go Report Card](https://goreportcard.com/badge/github.com/schemahero/schemahero?style=flat-square)](https://goreportcard.com/report/github.com/schemahero/schemahero)
[![Coverage](https://codecov.io/gh/schemahero/schemahero/branch/master/graph/badge.svg)](https://codecov.io/gh/schemahero/schemahero)
[![Build Status](https://badge.buildkite.com/deaf7798e8cc5f726c9684514a4e63285123481ee410aad94e.svg?branch=master)](https://buildkite.com/replicated/schemahero)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/schemahero/schemahero)
[![LICENSE](https://img.shields.io/github/license/schemahero/schemahero.svg?style=flat-square)](https://github.com/schemahero/schemahero/blob/master/LICENSE)

**Note**: This is a work-in-progress that is not yet functional. SchemaHero is a an experiment right now, and does not have enough implementation to be used in any environment. Work is in progress.

## What is SchemaHero?

SchemaHero is a Kubernetes Operator for Declarative Schema Management for various databases. SchemaHero has the following goals:

1. Database tables can be expressed as [Kubernetes resources](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_table.yaml) that can be updated and deployed to the cluster.
2. Database migrations can be written as SQL statements, expressed as [Kubernetes resources](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_migration.yaml) that can be deployed to the cluster.
3. Database schemas can be [monitored for drift](https://github.com/schemahero/schemahero/blob/master/config/samples/databases_v1alpha1_database.yaml) and brought back to the desired state automatically.
4. Schemas and migations can [require other schemas or migrations](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_table.yaml#L30) instead of ordering with timestamps and/or sequences.

## Getting Started

The recommended way to deploy SchemaHero is:

```
kubectl apply -f https://raw.githubusercontent.com/schemahero/schemahero/master/install/schemahero/schemahero-operator.yaml
```

To get started, read our [tutorial](https://github.com/schemahero/schemahero/blob/master/docs/tutorial) and the [how to use guide](https://github.com/schemahero/schemahero/blob/master/docs/how-to-use)


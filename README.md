# SchemaHero

[![Go Report Card](https://goreportcard.com/badge/github.com/schemahero/schemahero?style=flat-square)](https://goreportcard.com/report/github.com/schemahero/schemahero)
[![Coverage](https://codecov.io/gh/schemahero/schemahero/branch/master/graph/badge.svg)](https://codecov.io/gh/schemahero/schemahero)
[![Build Status](https://badge.buildkite.com/deaf7798e8cc5f726c9684514a4e63285123481ee410aad94e.svg?branch=master)](https://buildkite.com/replicated/schemahero)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/schemahero/schemahero)
[![LICENSE](https://img.shields.io/github/license/schemahero/schemahero.svg?style=flat-square)](https://github.com/schemahero/schemahero/blob/master/LICENSE)

## What is SchemaHero?

SchemaHero is a Kubernetes Operator for [Declarative Schema Management](https://schemahero.io/background/declarative-schema-management/) for [various databases](https://schemahero.io/databases/). SchemaHero has the following goals:

1. Database table schemas can be expressed as [Kubernetes resources](https://schemahero.io/how-to-use/deploying-tables/creating-tables/) that can be deployed to a cluster.
2. Database schemas can be edited and deployed to the cluster. SchemaHero will calculate the required change (`ALTER TABLE` statement) and apply it.
3. SchemaHero can manage databases that are deployed to the cluster, or external to the cluster (RDS, Google CloudSQL, etc).

## Getting Started

The recommended way to deploy SchemaHero is:

```
kubectl apply -f https://raw.githubusercontent.com/schemahero/schemahero/master/install/schemahero/schemahero-operator.yaml
```

To get started, read our [tutorial](https://github.com/schemahero/schemahero/blob/master/docs/tutorial) and the [how to use guide](https://github.com/schemahero/schemahero/blob/master/docs/how-to-use)


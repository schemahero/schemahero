<div align="center">
  <img alt="SchemaHero" src="./schemahero_logo.svg" width="600px" />
</div>
<br/>
<br/>

[![Go Report Card](https://goreportcard.com/badge/github.com/schemahero/schemahero?style=flat-square)](https://goreportcard.com/report/github.com/schemahero/schemahero)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/schemahero/schemahero)
[![LICENSE](https://img.shields.io/github/license/schemahero/schemahero.svg?style=flat-square)](https://github.com/schemahero/schemahero/blob/main/LICENSE)

## What is SchemaHero?

SchemaHero is a Kubernetes Operator for [Declarative Schema Management](https://schemahero.io/learn/declarative-schema-management/) for [various databases](https://schemahero.io/databases). SchemaHero has the following goals:

1. Database table schemas can be expressed as [Kubernetes resources](https://schemahero.io/docs/managing-tables/creating-tables/) that can be deployed to a cluster.
2. Database schemas can be edited and deployed to the cluster. SchemaHero will calculate the required change (`ALTER TABLE` statement) and apply it.
3. SchemaHero can manage databases that are deployed to the cluster, or external to the cluster (RDS, Google CloudSQL, etc).

## Getting Started

The recommended way to deploy SchemaHero is to use the kubectl plugin. Full instruction and other installation methods listed in the [documentation](https://schemahero.io/docs/installing/kubectl/).

To get started, read the [tutorial](https://schemahero.io/learn/tutorial/introduction) and the [full documentation](https://schemahero.io/docs/)

# Community

For questions about using SchemaHero, there's a [Replicated Community](https://help.replicated.com/community) forum, and a [#schemahero channel in Kubernetes Slack](https://kubernetes.slack.com/channels/schemahero). If you are interested in contributing to Schemahero, you can find more information on how to get started  on our [Community page](https://schemahero.io/community/).

If you're a user of SchemaHero, add yourself to the [ADOPTERS](/ADOPTERS) file!

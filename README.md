# SchemaHero

[![Build Status](https://travis-ci.org/schemahero/schemahero.svg?branch=master)](https://travis-ci.org/schemahero/schemahero)

**Note**: This is a work-in-progress that is not yet functional. SchemaHero is a an experiment right now, and does not have enough implementation to be used in any environment. Work is in progress.

## What is SchemaHero?

SchemaHero is a Kubernetes operator to manage database migrations and schemas as Kubernetes manifests.

1. Database tables can be expressed as [Kubernetes resources](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_table.yaml) that can be updated and deployed to the cluster.
2. Database migrations can be written as SQL statements, expressed as [Kubernetes resources](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_migration.yaml) that can be deployed to the cluster.
3. Database schemas can be [monitored for drift](https://github.com/schemahero/schemahero/blob/master/config/samples/databases_v1alpha1_database.yaml) and brought back to the desired state automatically.
4. Schemas and migations can [require other schemas or migrations](https://github.com/schemahero/schemahero/blob/master/config/samples/schemas_v1alpha1_table.yaml#L30) instead of ordering with timestamps and/or sequences.

## Getting Started

The recommended way to deploy SchemaHero is:

```
kubectl apply -f https://github.com/schemahero/schemahero/tree/master/install/k8s
```

If you need any customizations, use [Ship](https://github.com/replicatedhq/ship):

```
brew install ship
ship init github.com/schemahero/schemahero/tree/masterinstall/k8s
```

### Questions

*What about rollbacks?*
You can include the "delete" script. And that will be executed on `kubectl delete`.

*How can I control the order?*
The order that migrations run in is important. Each migration supports a "requires" field. This is the name (or names) of other migrations that must be applied before this one.

This will allow your developers to submit migrations in any order, but each developer can choose their "base".

SchemaHero will calculate a Directed Acylic Graph (DAG) to determine the optimal order to apply migrations, and will always honor the `requires` field. But it's possible to run multiple migrations simultaneously, which improves performance when bootstrapping new databases.

*How about non-sql migrations? I need to write some go code?*
Currently, you can write these types of schema as a container, and include that in the schema.


### Migration To SchemaHero

If you are using another tool (Goose, db-migrate, Flyaway, etc), and want to convert to the cloud-native, Kubernetes-first, SchemaHero, it's recommended that you "rebase" your current migrations by starting with a clean tree.

To do this, create an export of your schema, and install it as the first 1-n migrations.

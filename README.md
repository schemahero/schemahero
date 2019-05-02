[![Build Status](https://travis-ci.org/schemahero/schemahero.svg?branch=master)](https://travis-ci.org/schemahero/schemahero)

# Project Status

This is a work-in-progress that is not yet functional. SchemaHero is a an experiment right now, and does not have enough implementation to be used in any environment. Work is in progress.


# What is SchemaHero?

SchemaHero is a Kubernetes operator that will allow you to deploy database migration schemas as Kubernetes manifests.

For example:

```
apiVersion: schemas.schemahero.io/v1alpha1
kind: Migration
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: CreateMoviesTable
spec:
  requires: ["CreateActorsTable"]
  create: |
    create table movies (
      id text not null primary key,
      title text not null
    );

  delete: |
    drop table if exists movies;

```

Then, you can:

```shell
$ kubectl apply -f ./CreateMoviesTable

$ kubectl get migrations
---
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

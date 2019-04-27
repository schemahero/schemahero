# Postgres Integration Tests

This directory contains custom resources and validation to ensure that a postgres database can be initialized and created, tables and schema migrations can be applied using the SchemaHero operator.

To execute these tests, install [kind](https://github.com/kubernetes-sigs/kind) and then execute:

```
make test
```

Note, these tests are executed as part of the continuous integration process that runs from the top level Makefile.

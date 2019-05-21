# Migrating To SchemaHero

Like most of us, you probably already have a database (or several) running in production. You've invested in your current tooling and have a many database migrations. The process of converting these into YAML that can be managed with SchemaHero isn't difficult, but it's time consuming and tedious. SchemaHero has functionality that can help, if you can access a running instance of your database.

Note: We recommend that you run the import tool against a local, dev instance when possible.

## Schema Import

To start, get the `schemahero` binary (or use the Docker container) and provide a connection string to your database:

```
$ schemahero generate \
    --driver postgres \
    --uri postgres://user:pass@host:5432/dbname \
    --dbname destired-schemahero-databasename \
    --output-dir ./imported
```

or

```
$ docker run -v `pwd`/imported:/out \
    schemahero/schemahero:alpha \
    generate \
    --driver postgres \
    --uri postgres://user:password@host:5432/db?sslmode=disable \
    --dbname desired-schemahero-name \
    --output-dir /out
```

This will create .yaml files (1 per table) that you can deploy to a cluster to recreate the schema

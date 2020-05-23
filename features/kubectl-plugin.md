# SchemaHero kubectl plugin

## Summary

This proposal is to change the default behavior of the SchemaHero operator so that it doesn't deploy migrations when received. Instead, there should be a kubectl plugin that can be used to interact with the operator and approve or reject changes. This plugin should provide the ability to review the generated SQL statements. The Database kind should be extended to support disabling this functionality and automatically deploying, if desired.

## Motivation

Automatically deploying schema changes is risky because there's little or no opportunity to review the generated statements or to group / order schema changes.

Reviewing SQL statements is important and the lack of ability is a barrier to adoption of a tool like SchemaHero.

Some grouping and ordering can be achieved if manually deploying changes using kubectl apply, but if using an automated workflow, multiple and out of order changes could be deployed at the same time.

## Proposal

This is the current plan of action; it's based on product feedback received so far since launching SchemaHero.

- Bump the Table and Database CRD versions to v1alpha4
- Add a top level boolean attribute to the Database kind named `automaticallyDeployMigrations`
- When a table migration is deployed, if `automaticallyDeployMigrations` is not enabled, the controller should only generate the desired YAML
- The desired YAML should be stored in the Status field of the Table object


- Create a kubectl plugin named schemahero that interacts with the running operator using the kubecontext
- The kubectl plugin should be able to list pending migrations and view the SQL
- The plugin should be able to request that the generated SQL be regenerated from the current schema
- The plugin should be able to approve or reject the migrations, as desired

As an example to help make this easier to understand, a use of the schemahero plugin could look like this:

Some basic interactions:

1. List databases
```shell
$ kubectl schemahero get databases
NAME          PENDING
mydb          0
reporting     1
```

2. List tables in one database (or --all-databases should be supported)
```shell
$ kubectl schemahero get tables --database reporting
NAME          PENDING
users         0
projects      1
```

3. List all pending migrations for all database (--status=pending) should be default?
```shell
$ kubectl schemahero get migrations --status=pending --all-databases
ID                 DATABASE          TABLE
abc123             reporting         projects
```

4. View a migration, including the generated SQL
```shell
$ kubectl schemahero describe migration abc123
Name: abc123
Database:
      Name: reporting
      Type: postgres
      AutomaticallyDeployMigrations: false
Table:
      Columns:
            - Name: id
              Type: text


Status:
      Migration Created At: 2019-12-18T12:34:56Z
      Migration: UPDATE reporting ADD COLUMN something (text)
```

5. Regenerate a migration (recalculate the sql)
```shell
$ kubectl schemahero update migration abc123 --regenerate
Migration abc123 will be regenerate.
```


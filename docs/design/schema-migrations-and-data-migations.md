# Schema Migrations and Data Migations

There are two types of migratinos that have to be managed and deployed:

1. Schema Migrations
1. Data Migrations

## Schema Migrations

A Schema migration can be expressed in SQL syntax, and alters the structure of the database. These often are new tables, changing columns, altering indexed data and more. These are commonly written and can always be expressed in an idempotent syntax. Different database engines enforce various rules on how these can be applied. For example, MySQL will not allow a schema migration to be executed in a transaction, while Postgres will. Schema management is often unique to the database.

## Data Migrations

Less frequently, a developer must migrate some data to a new format in a database. This can involve calculating a new column and writing it, or creating new values in code and inserting them. Many traditional database management tools blend the tasks of schema migrations and data migrations into one tool.

When looking at adding a data migration to a project, there is often a way to achieve the same result by implementing the update differently.

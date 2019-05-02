# Desired State Database Schema Management

There are many benefits to managing database schemas as code, including:
1. Ability to adhere to a change management process
2. Repeatable deployments to new environments
...

## Current Tools

There are several commonly-used methods of managing database schemas employed today.

### Sequential Migrations / Replay

Tools such as db-migrate, Flyaway, goose and others fit into this category. These are immutable after deployment, stacked migrations that start from an empty database and consist of ordered create, alter and drop commands.

#### Challenges

These tools work nicely at first, but over time create some operational challenges:

1. When a feature is no longer used, the database runtime must continue to support this to allow this migration to succeed. For example, if using Postgres and an extension is required and used, and then removed, the database migations will fail to run on a new database unless that extension is present, even though it's not needed in the end state.

2. Performance starts to become slow on new environments. Eventually, in a rapid-iteration product, there can be hundreds of migrations. Replying these can be slow and any single failed migration will break the deployment.

3. Database upgrades create incompatible migrations. After upgrading a database version, the syntax supported may change. This can leave older migrations unable to be applied against the current version of the databse. 

4. Concurrent changes can create conflicts or skipped migrations. These tools often employ a sequential (integer) counter or a timestamp. When multiple migrations are simultaneously prepared offline, these may have the same counter value or be commited in a different order than they were generated. This can cause the runtime to skip a migration.

5. No dependency management between teams.

#### Workarounds

To help solve this, manual intervention is often taken to "rebase" the migrations. This is equivalent to retrieving a current schema from the database, deleting all migrations, and creating a single migration. This is a manual process that must be run occaisionally when using a sequential migration strategy.

#### Benefits

1. Ordering of columns is guaranteed. 
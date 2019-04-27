# Databases

Before deploying database tables and schemas, a database object should be deployed. The CRD for databases does two things:

1. It creates an object that can be referenced in tables and schemas.

This makes it easier to deploy schemas so that the credentials and database connection string doesn't have to be included in every schema migration that's deployed.

2. It creates a watch object in the cluster to watch the database.

When a database object is deployed, the operator will attempt to connect to it using the credentials provided. Each provider implements a "watch" that will be able to detect drift from the desired schema. When drift is detected, it will correct it to being it back into a managed state.


# Schemahero Integration Tests

These tests verify SchemaHero by:

1. Creating a database with an init script so that there are some predefined tables
2. Applying a table.yaml
3. Generating fixtures using schemahero
4. Verifying that the generated fixtures are correct


Ideally these tests should be executed from Go code for more reliability and easier to use.

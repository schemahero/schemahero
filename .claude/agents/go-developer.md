---
name: go-developer
description: Writes go code for this project
---

You are the agent that is invoked when needing to add or modify go code in this repo. 

* **Imports** - when importing local references, the import path is ALWAYS "github.com/schemahero/schemahero". 



* **SQL** - we write sql statements right in the code, not using any ORM. SchemaHero defined the schema, but there is no run-time ORM here and we don't want to introduce one.

* **ID Generation** -
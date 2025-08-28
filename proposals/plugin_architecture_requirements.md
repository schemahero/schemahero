Goal: Remove the "in-tree" database engine from the project in order to:
    1. make it easier to maintain so that database engine logic can be separated into separate codebases (projects, repos?)
    2. make it easier to add a new (even propietary) engine support 
    3. speed up testing and dev of database engines

Description:
We want to remove the in-tree implementation of mysql, postgres, cassandra, etc from this project, and make these separate plugins.
The schemahero process should be able to initialize and work with these dynamically.
We still need the operator to be able to install the plugins for some hardended environments that can't download plugins at runtime.
The plugins should conform to an interface and all be separately testable.
We don't care where the plugin is built: from the main schemahero repo (maybe a /plugins directory with separate go projects), a shared plugins, or even local.
Adding a plugin should probably specify the URL/release/version, and optionally a path. 
Officially supported plugins should come from a known location ("registry")
The current CRD/custom resource should not be aware of this change. We list the engine, if the engine plugin is available it should just run it, else download. Maybe an optional extension to the CRD to specify the path to the extension (in the Database kind) only to override the default behavior of looking in the default registry.

Some ideas / requirements:
    1. maybe we should look at hashicorp/go-plugin here. each engine can be a binary that is RPC communicating 
    2. schemahero itself should have a command to download the engine(s) and verify checksums
    3. do this with no breaking changes

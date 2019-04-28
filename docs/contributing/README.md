# Contributing

Contributions to SchemaHero are welcome! This document helps explain how to set up a local environment to run, test and validate a local copy before submitting a pull request. Because of the various databases and paths that SchemaHero supports, it would be a difficult task to manaually test all of the supported configurations. Therefore, we rely heavily on automation tests that are in the /integration directory. If it's not tested, it's not guaranteed to work.

## Local Environment

It's helpful to be able to run a local environment, with a database, and be able to apply schema changes using an edited copy of this code. This is easy to do, if you have Docker and Kubernetes installed.

(These docs were written from a setup using Docker for Mac, with the Docker-provided Kubernetes installed. The ideas should work for Windows and for linux-based installations such as microk8s, but these steps have not been verified in those environments yet).

### Build

After cloning the repo, run `make` to build the binaries and execute the unit tests. This will not execute the integration tests. Integration tests are very important here, but they can take 30+ minutes to run, so they are set up to run on demand (and in CI).

### Running

To run a local copy, `make install run`. This will run the manager on a local workstation, but will connect to the cluster. Kubebuilder is providing this framework and plumbing. After the first time, `make run` is enough to keep this updated.

### A Local Database

Running locally means you'll want a local database. you can `kubectl apply -f ./config/dev/database/postgres.yaml` to run a local postgres in your cluster. This will create a statefulset. To start clean, delete the PVC that is automatically provisioned.

### Working Locally

Now that the manager is running, it's ready to receive a custom resource. Deploy a `database` type by opening a new shell and typing `kubectl apply -f ./config/dev/github/pg.yaml`


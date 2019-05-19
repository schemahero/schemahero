# SchemaHero Release Strategy

There are 2 images that are used in the SchemaHero Operator:

schemahero/schemahero and `schemahero/schemahero-manager`.

The -manager image is the controller and manager for the operator. It runs in the cluster and handles the reconciliation of any deployed custom resources. When a database is deployed, the code to connect and monitor the connection is in the schemahero container. Both of these container images are built from this repo.

The reason for this separation is to allow an easy way to run a schemahero binary also. This is useful in dev environment and migrations.

These two images are tagged and released at the same time.

## Alpha

There is an `:alpha` tag of these iamges. This is the latest commit to master. It may or may not be stable. It's not recommded to run `:alpha` on a production system.

## Latest

The `:latest` image points to the last stable release of the images.

## major.minor.patch

The `:x.y.z` tag points to a specific, immutable revision. These are created when a tag is pushed. These are the most stable versions of SchemaHero and recommended to use in production.


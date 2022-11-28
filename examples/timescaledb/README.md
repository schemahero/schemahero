# TimescaleDB example

This example was designed to use in dev when validating or improving SchemaHero TimescaleDB support.

To use:

1. Deploy the `timescale.yaml` to a cluster.

```
kubectl apply -f ./timescale.yaml
```

2. Forward "localhost:5432" to the service:

```
kubectl port-forward svc/timescale 5432:5432
```

3. Run a local SchemaHero with:

```
make install run
```

4. Deploy the database:

```
kubectl apply -f ./airlinedb.yaml
```

5. Deploy tables

```
kubectl apply -f ./tables
```

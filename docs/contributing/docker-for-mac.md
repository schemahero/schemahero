# Docker For Mac

It's possible to completely run SchemaHero on a Docker For Mac setup. To do so:

### Enable Kubernetes in Docker For Mac

### Add some hostname trickery

Some of the components will connect in-cluster, but the manager connects from out. To faciliate this, we can create a hosts file entry on the Mac to handle what happens automatically in cluster:
```
echo "127.0.0.1 postgresql" | sudo tee -a /etc/hosts > /dev/null
```

### Run the schemahero manager outside of the cluster
```
make install run
```

### Start a port forward

Port forward localhost:5432 to the cluster ip running in the cluster:

```
kubectl port-forward svc/postgresql 5432:5432
```

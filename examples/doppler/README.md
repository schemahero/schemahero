# Doppler database connection

Store the database connection string in a Doppler secret such as `DATABASE_URL`.
Create a read-only Doppler service token scoped to the project and config, then
store that token in the same Kubernetes namespace as the SchemaHero `Database`:

```shell
kubectl create secret generic doppler-credentials \
  --namespace production \
  --from-literal=token='dp.st.prd.example'
```

Update the project and config in `database.yaml`, then apply it:

```shell
kubectl apply -f database.yaml
```

SchemaHero reads the token from the Kubernetes Secret and requests the computed
value of `DATABASE_URL` from Doppler whenever it resolves the database
connection. Neither the token nor the resolved connection string is stored in
the `Database` resource.

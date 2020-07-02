## SchemaHero with Vault

The files in this directory are an example of using SchemaHero to manage a PostgreSQL database using credentials managed by HashiCorp Vault.

### Deploy Postgres and Vault

Deploy a PostgreSQL and Vault instance to a new namespace. 
The following manifests were taken from the PostgreSQL and Vault Helm Charts.
The vault chart is in dev mode and should be used for this tutorial only -- do not use for production.

```
kubectl create ns schemahero-vault
kubectl apply -f ./postgresql/postgresql-11.8.0.yaml
kubectl apply -f ./vault/vault.yaml
```

### Configure Vault


Next, enable the database secret engine:

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault secrets enable database
```

Now, we need to create a Vault role and config:

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault write database/roles/schemahero \
    db_name=airlinedb \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    revocation_statements="ALTER ROLE \"{{name}}\" NOLOGIN;"\
    default_ttl="1h" \
    max_ttl="24h"
```

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault write database/config/airlinedb \
    plugin_name=postgresql-database-plugin \
    allowed_roles="*" \
    connection_url="postgresql://{{username}}:{{password}}@postgresql:5432/airlinedb?sslmode=disable" \
    username="postgres" \
    password="password"
```

### Verify vault is working

The following command will request a new username and password for our database.
This is just confirming that Vault it working and has permissions.

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault read database/creds/schemahero
```

## Enable Kubernetes auth in Vault

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault auth enable kubernetes
```

FROM YOUR COMPUTER (not in the vault pod):

```
kubectl -n schemahero-vault exec $(kubectl -n schemahero-vault get pods --selector "app.kubernetes.io/instance=vault,component=server" -o jsonpath="{.items[0].metadata.name}") -c vault -- \
  sh -c ' \
    VAULT_TOKEN=root vault write auth/kubernetes/config \
       token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
       kubernetes_host=https://${KUBERNETES_PORT_443_TCP_ADDR}:443 \
       kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
```

Create the policy

```
tee -a /tmp/policy.hcl > /dev/null <<EOT
path "database/creds/schemahero" {
  capabilities = ["read"]
}
path "database/config/airlinedb" {
  capabilities = ["read"]
}
EOT
```

```
kubectl -n schemahero-vault cp /tmp/policy.hcl vault-0:/tmp/policy.hcl
```


```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault policy write schemahero /tmp/policy.hcl
```

```
kubectl exec -n schemahero-vault -it vault-0 -- env VAULT_TOKEN=root vault write auth/kubernetes/role/schemahero \
    bound_service_account_names=schemahero \
    bound_service_account_namespaces=schemahero-vault \
    policies=schemahero \
    ttl=1h
```

Deploy the serviceaccount:


```
kubectl apply -f ./vault/sa.yaml
```

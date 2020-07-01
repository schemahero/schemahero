To generate the vault.yaml:

helm template vault hashicorp/vault --namespace schemahero-vault --set server.dev.enabled=true
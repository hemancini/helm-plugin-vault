# helm vault

A Helm plugin for managing VaultStaticSecret resources.

## Install

```sh
helm plugin install https://github.com/hemancini/helm-plugin-vault.git
```

## Commands

- `helm vault get-common -ns <NAMESPACE>`: Output a VaultStaticSecret resource in YAML format
- `helm vault upgrade-common <DEPLOYMENT> -ns <NAMESPACE>`: Install or upgrade a VaultStaticSecret common-secrets resource
- `helm vault delete-common <DEPLOYMENT> -ns <NAMESPACE>`: Delete an installed VaultStaticSecret common-secrets resource

## Uninstall

```sh
helm plugin uninstall vault
```

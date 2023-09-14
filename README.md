# helm vault

A Helm plugin for managing VaultStaticSecret resources.

## Install

```sh
helm plugin install https://github.com/hemancini/helm-plugin-vault.git
```

## Commands

- `helm vault get -n <NAMESPACE>`: Output a VaultStaticSecret resource in YAML format
- `helm vault delete <DEPLOYMENT> -n <NAMESPACE>`: Delete an installed VaultStaticSecret common-secrets resource
- `helm vault upgrade <DEPLOYMENT> -n <NAMESPACE> -f <HELM_VALUES>`: Install or upgrade a VaultStaticSecret common-secrets resource
- `helm vault help`: Help about any command

## Uninstall

```sh
helm plugin uninstall vault
```

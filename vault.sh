#!/bin/bash

set -e

usage() {
    cat <<EOF
Available Commands:
    helm vault get-common -ns <NAMESPACE>                        Output a VaultStaticSecret resource in YAML format  
    helm vault upgrade-common <DEPLOYMENT> -ns <NAMESPACE>       Install or upgrade a VaultStaticSecret common-secrets resource
    helm vault delete-common <DEPLOYMENT> -ns <NAMESPACE>        Delete an installed VaultStaticSecret common-secrets resource
    --help                                                       Display this text
EOF
}

# Create the passthru array
PASSTHRU=()
HELP=FALSE

K8S_NAMESPACE=""
K8S_DEPLOYMENT=""
REPO="common-secrets"

while [[ $# -gt 0 ]]; do
    case "$1" in
    upgrade-common | delete-common)
        K8S_DEPLOYMENT="$2"
        PASSTHRU+=("$1")
        shift
        ;;
    -ns)
        K8S_NAMESPACE="$2"
        shift
        ;;
    --help)
        HELP=TRUE
        shift # past argument
        ;;
    *)                   # unknown option
        PASSTHRU+=("$1") # save it in an array for later
        shift            # past argument
        ;;
    esac
done

# Restore PASSTHRU parameters
set -- "${PASSTHRU[@]}"

# Show help if flagged
if [ "$HELP" == "TRUE" ]; then
    usage
    exit 0
fi

if [ "$K8S_NAMESPACE" == "" ]; then
    echo -e "Parameter -ns is required.\n"
    usage
    exit 1
fi

if [[ "$K8S_DEPLOYMENT" == "-ns" || "$K8S_DEPLOYMENT" == "" ]]; then
    echo -e "Parameter <DEPLOYMENT> is required.\n"
    usage
    exit 1
fi

getVaultStaticSecret() {
    K_GET="$(kubectl get vaultstaticsecret "$REPO" -n "$K8S_NAMESPACE" -o yaml)"
    echo "$K_GET"
}

# COMMAND must be either 'get-common', 'upgrade-common', or 'delete-common'
COMMAND=${PASSTHRU[0]}

if [ "$COMMAND" == "get-common" ]; then
    getVaultStaticSecret | yq
    exit 0
elif [ "$COMMAND" == "upgrade-common" ]; then
    VAULT_STATIC_SECRET="$(getVaultStaticSecret)"
    echo "$VAULT_STATIC_SECRET" >"$REPO.yaml"

    deploy_exists=$(yq eval '.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'")' "$REPO.yaml")
    if [ "$deploy_exists" == "" ]; then
        echo "$K8S_DEPLOYMENT does not exist. Adding it."
        yq eval '(.spec.rolloutRestartTargets += [{"kind": "Deployment", "name": "'"${K8S_DEPLOYMENT}"'"}])' -i "$REPO.yaml"
        cat $REPO.yaml | kubectl apply -n "$K8S_NAMESPACE" -f -
    else
        echo "$K8S_DEPLOYMENT already exists."
        exit 0
    fi

    yq '.spec.rolloutRestartTargets[].name' "$REPO.yaml"
    exit 0
elif [ "$COMMAND" == "delete-common" ]; then
    VAULT_STATIC_SECRET="$(getVaultStaticSecret)"
    echo "$VAULT_STATIC_SECRET" >"$REPO.yaml"

    deploy_exists=$(yq eval '.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'")' "$REPO.yaml")
    if [ -n "$deploy_exists" ]; then
        echo "$K8S_DEPLOYMENT exists. Deleting it."
        yq eval 'del(.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'"))' -i "$REPO.yaml"
        cat $REPO.yaml | kubectl apply -n "$K8S_NAMESPACE" -f -
    else
        echo "$K8S_DEPLOYMENT not exists."
        exit 0
    fi

    yq '.spec.rolloutRestartTargets[].name' "$REPO.yaml"
    exit 0
else
    echo "Error: Invalid command."
    usage
    exit 1
fi

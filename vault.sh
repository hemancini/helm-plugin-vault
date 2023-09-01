#!/bin/bash

set -e

if ! command -v yq &>/dev/null; then
    echo "yq not installed in the system."
    exit 1
fi

usage() {
    cat <<EOF
Available Commands:
    helm vault get-common -ns <NAMESPACE>                                       Output a VaultStaticSecret resource in YAML format  
    helm vault upgrade-common <DEPLOYMENT> -ns <NAMESPACE> -f <HELM_VALUES>     Install or upgrade a VaultStaticSecret common-secrets resource
    helm vault delete-common <DEPLOYMENT> -ns <NAMESPACE>                       Delete an installed VaultStaticSecret common-secrets resource
    --help                                                                      Display this text
EOF
}

# Create the passthru array
PASSTHRU=()
HELP=FALSE

K8S_NAMESPACE=""
K8S_DEPLOYMENT=""
HELM_VALUES=""
RESOURCE="common-secrets"

while [[ $# -gt 0 ]]; do
    case "$1" in
    upgrade-common | delete-common)
        K8S_DEPLOYMENT="$2"
        PASSTHRU+=("$1")
        shift 2
        ;;
    -ns)
        K8S_NAMESPACE="$2"
        shift 2
        ;;
    -f)
        HELM_VALUES="$2"
        shift 2
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
else
    if [ -e "$HELM_VALUES" ]; then
        deploy_exists=$(yq eval '.deployment.commonSecrets[]' "$HELM_VALUES")
        if [ -n "$deploy_exists" ]; then
            echo "Common secrets availables..."
        else
            echo "Common secrets not availables."
            exit 0
        fi
    else
        echo "The file $HELM_VALUES does not exist."
        exit 1
    fi
fi

if [[ "$K8S_DEPLOYMENT" == "-ns" || "$K8S_DEPLOYMENT" == "" ]]; then
    echo -e "Parameter <DEPLOYMENT> is required.\n"
    usage
    exit 1
fi

getVaultStaticSecret() {
    K_GET="$(kubectl get vaultstaticsecret "$RESOURCE" -n "$K8S_NAMESPACE" -o yaml)"
    echo "$K_GET"
}

clear() {
    rm -rf "$RESOURCE.yaml"
    exit 0
}

# COMMAND must be either 'get-common', 'upgrade-common', or 'delete-common'
COMMAND=${PASSTHRU[0]}

if [ "$COMMAND" == "get-common" ]; then
    getVaultStaticSecret | yq
    exit 0
elif [ "$COMMAND" == "upgrade-common" ]; then
    VAULT_STATIC_SECRET="$(getVaultStaticSecret)"
    echo "$VAULT_STATIC_SECRET" >"$RESOURCE.yaml"

    deploy_exists=$(yq eval '.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'")' "$RESOURCE.yaml")
    if [ "$deploy_exists" == "" ]; then
        echo "$K8S_DEPLOYMENT does not exist. Adding it."
        yq eval '(.spec.rolloutRestartTargets += [{"kind": "Deployment", "name": "'"${K8S_DEPLOYMENT}"'"}])' -i "$RESOURCE.yaml"
        cat $RESOURCE.yaml | kubectl apply -n "$K8S_NAMESPACE" -f -
    else
        echo "$K8S_DEPLOYMENT already exists."
        clear
    fi

    yq '.spec.rolloutRestartTargets[].name' "$RESOURCE.yaml"
    clear
elif [ "$COMMAND" == "delete-common" ]; then
    VAULT_STATIC_SECRET="$(getVaultStaticSecret)"
    echo "$VAULT_STATIC_SECRET" >"$RESOURCE.yaml"

    deploy_exists=$(yq eval '.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'")' "$RESOURCE.yaml")
    if [ -n "$deploy_exists" ]; then
        echo "$K8S_DEPLOYMENT exists. Deleting it."
        yq eval 'del(.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'"))' -i "$RESOURCE.yaml"
        cat $RESOURCE.yaml | kubectl apply -n "$K8S_NAMESPACE" -f -
    else
        echo "$K8S_DEPLOYMENT not exists."
        exit 0
    fi

    yq '.spec.rolloutRestartTargets[].name' "$RESOURCE.yaml"
    clear
else
    echo "Error: Invalid command."
    usage
    exit 1
fi

#!/bin/bash

set -e

if ! command -v yq &>/dev/null; then
    echo "yq not installed in the system."
    exit 1
fi

usage() {
    cat <<EOF
Available Commands:
    helm vault get-common -n <NAMESPACE>                                       Output a VaultStaticSecret resource in YAML format  
    helm vault upgrade-common <DEPLOYMENT> -n <NAMESPACE> -f <HELM_VALUES>     Install or upgrade a VaultStaticSecret common-secrets resource
    helm vault delete-common <DEPLOYMENT> -n <NAMESPACE>                       Delete an installed VaultStaticSecret common-secrets resource
    --help                                                                     Display this text
EOF
}

helmValues() {
    if [ "$HELM_VALUES" == "" ]; then
        echo -e "Parameter -f is required.\n"
        usage
        exit 1
    fi
}

helmValuesCommonSecrets() {
    if [ -e "$HELM_VALUES" ]; then
        deploy_exists=$(yq eval '.deployment.commonSecrets[]' "$HELM_VALUES")
        if [ -n "$deploy_exists" ]; then
            echo "Common secrets availables..."
        else
            echo "Common secrets not availables."
            exit 0
        fi
    else
        echo "The file \"$HELM_VALUES\" does not exist."
        exit 1
    fi
}

getDeploymentName() {
    if [[ "$K8S_DEPLOYMENT" == "" ]]; then
        echo -e "Parameter <DEPLOYMENT> is required.\n"
        usage
        exit 1
    fi
}

getValutEnvironment() {
    V_ENVIRONMENT="$(yq eval '.vaultSecrets.environment' "$HELM_VALUES")"
    echo "$V_ENVIRONMENT"
}

getValutEnabled() {
    V_ENABLED="$(yq eval '.vaultSecrets.enabled' "$HELM_VALUES")"
    echo "$V_ENABLED"
}

createVaultStaticSecret() {
    VAULT_PATH_ENVIRONMENT="$(getValutEnvironment)"
    if [ "$VAULT_PATH_ENVIRONMENT" == "null" ]; then
        echo "Error vault environment not found."
        exit 1
    fi
    cat <<EOF | kubectl apply -n "$K8S_NAMESPACE" -f - >/dev/null
apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultStaticSecret
metadata:
    labels:
        app.kubernetes.io/managed-by: Vault
    name: $RESOURCE
    namespace: $K8S_NAMESPACE
spec:
    destination:
        create: true
        name: $RESOURCE
    hmacSecretData: true
    mount: core
    path: $VAULT_PATH_ENVIRONMENT/$RESOURCE
    refreshAfter: 30s
    type: kv-v2
EOF
}

getVaultStaticSecret() {
    K_GET="$(kubectl get vaultstaticsecret "$RESOURCE" -n "$K8S_NAMESPACE" -o yaml 2>/dev/null)"
    if [ "$K_GET" == "" ]; then
        createVaultStaticSecret
        K_GET="$(kubectl get vaultstaticsecret "$RESOURCE" -n "$K8S_NAMESPACE" -o yaml 2>/dev/null)"
    fi
    echo "$K_GET"
}

getCommonSecrets() {
    K_GET="$(kubectl get secrets "$RESOURCE" -n "$K8S_NAMESPACE" -o yaml 2>/dev/null)"
    echo "$K_GET"
}

getValuesSecrets() {
    K_SECRETS="$(yq eval '.deployment.secrets' "$HELM_VALUES")"
    echo "$K_SECRETS"
}

# Input:
#   $1: secrets, $2: commonSecrets
# Output:
#   array diff
getNewCommonSecrets() {
    newCommonSecrets=()
    for secret in $1; do
        for commonSecret in $2; do
            if [ "$secret" == "$commonSecret" ]; then
                newCommonSecrets+=("$secret")
            fi
        done
    done
    echo "${newCommonSecrets[@]}"
}

# Input:
#   $1: secrets, $2: commonSecrets
# Output:
#   array diff
getNewSecrets() {
    newSecrets=()
    for secret in $1; do
        encontrado=false
        for elemento_B in $2; do
            if [[ "$secret" == "$elemento_B" ]]; then
                encontrado=true
                break
            fi
        done
        if [[ "$encontrado" == false ]]; then
            newSecrets+=("$secret")
        fi
    done
    echo "${newSecrets[@]}"
}

# Input:
#   $1: secrets, $2: commonSecrets
# Output:
#   void
updateValuesSecrets() {
    _secretsIndex=0
    _commonSecretsIndex=0

    for secret in $1; do
        if [ "$_secretsIndex" == "0" ]; then
            yq eval ".deployment.secrets = [\"$secret\"]" -i "$HELM_VALUES"
        else
            yq eval ".deployment.secrets += [\"$secret\"]" -i "$HELM_VALUES"
        fi
        _secretsIndex+=1
    done
    for commonSecret in $2; do
        if [ "$_commonSecretsIndex" == "0" ]; then
            yq eval ".deployment.commonSecrets = [\"$commonSecret\"]" -i "$HELM_VALUES"
        else
            yq eval ".deployment.commonSecrets += [\"$commonSecret\"]" -i "$HELM_VALUES"
        fi
        _commonSecretsIndex+=1
    done
}

checkVault() {
    VAULT_PATH_ENVIRONMENT="$(getValutEnvironment)"
    if [ "$VAULT_PATH_ENVIRONMENT" == "null" ]; then
        echo "vault environment not found."
        exit 0
    fi
    VAULT_ENABLED="$(getValutEnabled)"
    if [ "$VAULT_ENABLED" != "true" ]; then
        echo "vault is not enabled."
        exit 0
    fi
}

clear() {
    rm -rf "$RESOURCE.yaml"
    # sed sed -i '/commonSecrets:/d' "$HELM_VALUES"
    exit 0
}

# Create the passthru array
PASSTHRU=()
HELP=FALSE

RESOURCE="common-secrets"
K8S_NAMESPACE="$HELM_NAMESPACE"
K8S_DEPLOYMENT=""
HELM_VALUES=""
VAULT_PATH_ENVIRONMENT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
    upgrade-common | upgrade | delete-common)
        K8S_DEPLOYMENT="$2"
        PASSTHRU+=("$1")
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

# COMMAND must be either 'get-common', 'upgrade-common', or 'delete-common'
COMMAND=${PASSTHRU[0]}

if [ "$COMMAND" = "get-common" ] || [ "$COMMAND" = "get" ]; then
    getVaultStaticSecret | yq
    exit 0
elif [ "$COMMAND" == "upgrade-common" ] || [ "$COMMAND" == "upgrade" ]; then
    getDeploymentName # Check if deployment name is not empty
    helmValues     # Check if helm values file exists and common secrets are availables
    checkVault     # Check if vault is enabled and vault environment exists

    vaultStaticSecret="$(getVaultStaticSecret)"
    echo "$vaultStaticSecret" >"$RESOURCE.yaml"

    secrets="$(getValuesSecrets | sed 's/- //g')"
    commonSecrets="$(getCommonSecrets | yq eval '.data | keys' 2>/dev/null | sed 's/- //g')"

    newSecrets=$(getNewSecrets "$secrets" "$commonSecrets")
    newCommonSecrets=$(getNewCommonSecrets "$secrets" "$commonSecrets")
    updateValuesSecrets "$newSecrets" "$newCommonSecrets" # update chart values

    echo "environment: $VAULT_PATH_ENVIRONMENT"
    echo "secrets: [$(echo $newSecrets | sed 's/ /, /g')]"
    echo "common secrets: [$(echo $newCommonSecrets | sed 's/ /, /g')]"

    helmValuesCommonSecrets # read common secrets
    deploy_exists=$(yq eval '.spec.rolloutRestartTargets[] | select(.kind == "Deployment" and .name == "'"${K8S_DEPLOYMENT}"'")' "$RESOURCE.yaml")
    if [ "$deploy_exists" == "" ]; then
        echo "$K8S_DEPLOYMENT does not exist. Adding it."
        yq eval '(.spec.rolloutRestartTargets += [{"kind": "Deployment", "name": "'"${K8S_DEPLOYMENT}"'"}])' -i "$RESOURCE.yaml"
        cat $RESOURCE.yaml | kubectl apply -n "$K8S_NAMESPACE" -f -
    else
        echo "$K8S_DEPLOYMENT already exists in common-secrets."
        clear
    fi

    yq '.spec.rolloutRestartTargets[].name' "$RESOURCE.yaml" 2>/dev/null
    clear
elif [ "$COMMAND" == "delete-common" ] || [ "$COMMAND" == "delete" ]; then
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

    yq '.spec.rolloutRestartTargets[].name' "$RESOURCE.yaml" 2>/dev/null
    clear
else
    echo "Error: Invalid command."
    usage
    exit 1
fi

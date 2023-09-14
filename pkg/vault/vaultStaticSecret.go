package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"helm-plugin-vault/pkg/types"
	"helm-plugin-vault/pkg/utils"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	yamlsigs "sigs.k8s.io/yaml"
)

type VaultStaticSecret types.VaultStaticSecret
type ChartValues types.ChartValues
type Secret types.Secrets

var resourceCommonName = types.ResourceCommonName
var commonArray = utils.CommonArray
var diffArrays = utils.DiffArrays

func GetVaultCommonSecrets(clientset *kubernetes.Clientset, namespace string, showError bool) *VaultStaticSecret {
	vaultstaticsecrets, err := clientset.RESTClient().
		Get().
		AbsPath(types.ResourceCommonAbsPath).
		Resource(types.ResourceCommon).
		Name(resourceCommonName).
		Namespace(namespace).
		SetHeader("Accept", "application/yaml").
		DoRaw(context.TODO())

	if err != nil {
		if showError {
			fmt.Printf("Error: %v\n", err)
		}
		return nil
	}

	var vaultStaticSecret VaultStaticSecret
	if err := yaml.Unmarshal(vaultstaticsecrets, &vaultStaticSecret); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}
	return &vaultStaticSecret
}

func NewVaultCommonSecret(namespace string, filePath string) VaultStaticSecret {
	resourceName := resourceCommonName
	environment := getEnvironment(filePath)

	var vaultStaticSecret VaultStaticSecret
	vaultStaticSecret.APIVersion = "secrets.hashicorp.com/v1beta1"
	vaultStaticSecret.Kind = "VaultStaticSecret"
	vaultStaticSecret.Metadata.Name = resourceName
	vaultStaticSecret.Metadata.Namespace = namespace
	vaultStaticSecret.Metadata.Labels.AppKubernetesIoManagedBy = "vault"
	vaultStaticSecret.Spec.Mount = "core"
	vaultStaticSecret.Spec.Path = environment + "/" + resourceName
	vaultStaticSecret.Spec.Type = "kv-v2"
	vaultStaticSecret.Spec.RefreshAfter = "5m"
	vaultStaticSecret.Spec.Destination.Create = true
	vaultStaticSecret.Spec.Destination.Name = resourceName
	vaultStaticSecret.Spec.HmacSecretData = true
	vaultStaticSecret.Spec.RolloutRestartTargets = []struct {
		Kind string `yaml:"kind"`
		Name string `yaml:"name"`
	}{}

	return vaultStaticSecret
}

func CreateVaultCommonSecret(clientset *kubernetes.Clientset, namespace string, filePath string) *VaultStaticSecret {
	vaultStaticSecret := NewVaultCommonSecret(namespace, filePath)
	vaultStaticSecretYaml, err := yaml.Marshal(vaultStaticSecret)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	vaultStaticSecretJson, err := yamlsigs.YAMLToJSON(vaultStaticSecretYaml)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = clientset.RESTClient().
		Post().
		AbsPath(types.ResourceCommonAbsPath).
		Resource(types.ResourceCommon).
		Namespace(namespace).
		Body(vaultStaticSecretJson).
		DoRaw(context.TODO())

	if err != nil {
		log.Fatalf("error: %v", err)
	} else {
		fmt.Printf("Successfully created %s\n", resourceCommonName)
	}

	return &vaultStaticSecret
}

func PatchCommonSecrets(clientset *kubernetes.Clientset, namespace string, rolloutRestartTarget []types.RolloutRestartTarget) []byte {
	patchValue, err := json.Marshal(rolloutRestartTarget)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	patch := []byte(fmt.Sprintf(`[
		{"op": "add", "path": "/spec/rolloutRestartTargets", "value": %v}
	]`, string(patchValue)))

	data, err := clientset.RESTClient().
		Patch(k8sTypes.JSONPatchType).
		AbsPath(types.ResourceCommonAbsPath).
		Resource(types.ResourceCommon).
		Namespace(namespace).
		Name(types.ResourceCommonName).
		Body(patch).
		DoRaw(context.TODO())

	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf("Successfully patched %s\n", resourceCommonName)
	}

	return data
}

func DeleteCommonSecrets(clientset *kubernetes.Clientset, namespace string) {
	_, err := clientset.RESTClient().
		Delete().
		AbsPath(types.ResourceCommonAbsPath).
		Resource(types.ResourceCommon).
		Namespace(namespace).
		Name(types.ResourceCommonName).
		DoRaw(context.TODO())

	if err == nil {
		fmt.Printf("Successfully deleted %s\n", resourceCommonName)
	} else {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

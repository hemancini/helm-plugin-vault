package types

import "time"

var ResourceCommon = "vaultstaticsecrets"
var ResourceCommonName = "common-secrets"
var ResourceCommonAbsPath = "/apis/secrets.hashicorp.com/v1beta1"

type VaultStaticSecret struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Labels struct {
			AppKubernetesIoManagedBy string `yaml:"app.kubernetes.io/managed-by"`
		} `yaml:"labels"`
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Destination struct {
			Create bool   `yaml:"create"`
			Name   string `yaml:"name"`
		} `yaml:"destination"`
		HmacSecretData        bool   `yaml:"hmacSecretData"`
		Mount                 string `yaml:"mount"`
		Path                  string `yaml:"path"`
		RefreshAfter          string `yaml:"refreshAfter"`
		RolloutRestartTargets []struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		} `yaml:"rolloutRestartTargets"`
		Type string `yaml:"type"`
	} `yaml:"spec"`
}

type RolloutRestartTarget struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type Secrets struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Data       map[string]interface{} `yaml:"data"`
	Metadata   struct {
		CreationTimestamp time.Time `yaml:"creationTimestamp"`
		Labels            struct {
			AppKubernetesIoComponent          string `yaml:"app.kubernetes.io/component"`
			AppKubernetesIoManagedBy          string `yaml:"app.kubernetes.io/managed-by"`
			AppKubernetesIoName               string `yaml:"app.kubernetes.io/name"`
			SecretsHashicorpComVsoOwnerRefUID string `yaml:"secrets.hashicorp.com/vso-ownerRefUID"`
		} `yaml:"labels"`
		Name            string `yaml:"name"`
		Namespace       string `yaml:"namespace"`
		OwnerReferences []struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
			Name       string `yaml:"name"`
			UID        string `yaml:"uid"`
		} `yaml:"ownerReferences"`
		ResourceVersion string `yaml:"resourceVersion"`
		UID             string `yaml:"uid"`
	} `yaml:"metadata"`
	Type string `yaml:"type"`
}

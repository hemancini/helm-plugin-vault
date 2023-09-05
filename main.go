package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type VaultStaticSecret struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Annotations struct {
			HelmShResourcePolicy string `yaml:"helm.sh/resource-policy"`
			// KubectlKubernetesIoLastAppliedConfiguration string `yaml:"kubectl.kubernetes.io/last-applied-configuration"`
			MetaHelmShReleaseName      string `yaml:"meta.helm.sh/release-name"`
			MetaHelmShReleaseNamespace string `yaml:"meta.helm.sh/release-namespace"`
		} `yaml:"annotations"`
		// CreationTimestamp time.Time `yaml:"creationTimestamp"`
		// Generation        int       `yaml:"generation"`
		Labels struct {
			AppKubernetesIoManagedBy string `yaml:"app.kubernetes.io/managed-by"`
		} `yaml:"labels"`
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		// ResourceVersion string `yaml:"resourceVersion"`
		//	UID             string `yaml:"uid"`
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
	// Status struct {
	// 	SecretMAC string `yaml:"secretMAC"`
	// } `yaml:"status"`
}

type ChartValues struct {
	Deployment struct {
		ContainerPort int `yaml:"containerPort"`
		Envs          []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"envs"`
		Secrets       []string `yaml:"secrets"`
		CommonSecrets []string `yaml:"commonSecrets"`
	} `yaml:"deployment"`
	VaultSecrets struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"vaultSecrets"`
	Image struct {
		Repository string `yaml:"repository"`
		Tag        string `yaml:"tag"`
		PullPolicy string `yaml:"pullPolicy"`
	} `yaml:"image"`
	Namespace string `yaml:"namespace"`
	Ingress   struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"ingress"`
	LivenessProbe struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"livenessProbe"`
	ReadinessProbe struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"readinessProbe"`
}

var namespace string
var chartValues string

func main() {

	var rootCmd = &cobra.Command{Use: "vault"}
	clientset := k8sClient()

	var getCommon = &cobra.Command{
		Use:     "get-common",
		Short:   "Get common-secrets from vaultStaticSecret",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			vaultstaticsecrets := getCommonSecrets(clientset, namespace)
			fmt.Println(vaultstaticsecrets)
		},
	}

	var upgradeCommon = &cobra.Command{
		Use:     "upgrade-common",
		Short:   "Upgrade common-secrets from vaultStaticSecret",
		Aliases: []string{"upgrade"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("helm vault upgrade", args[0], "-n", namespace, "-f", chartValues)

			deployment := args[0]
			config := readYamlFile(chartValues)
			configSecrets := config.Deployment.Secrets
			vaultStaticSecret := getCommonSecrets(clientset, namespace)

			vaultStaticSecretUpdated := getUpsertCommonSecrets(configSecrets, vaultStaticSecret, deployment)
			vaultStaticSecretUpdated.Metadata.Name = "common-secrets"

			yaml, err := json.Marshal(&vaultStaticSecretUpdated)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			fmt.Printf("---\n%s\n\n", string(yaml))

			data, err := clientset.RESTClient().
				Post().
				// AbsPath("/apis/secrets.hashicorp.com/v1beta1/namespaces/bff-integration/vaultstaticsecrets/common-secrets").
				AbsPath("/apis/secrets.hashicorp.com/v1beta1").
				Resource("vaultstaticsecrets").
				Namespace(namespace).
				Body(yaml).
				DoRaw(context.Background())

			if err != nil {
				panic(err.Error())
			}

			fmt.Println(string(data))

			// Get().
			// AbsPath("/apis/secrets.hashicorp.com/v1beta1").
			// Resource("vaultstaticsecrets").
			// Namespace(namespace).
			// Name("common-secrets").
			// SetHeader("Accept", "application/yaml").
			// DoRaw(context.TODO())

		},
	}

	var deleteCommon = &cobra.Command{
		Use:     "delete-common",
		Short:   "Delete common-secrets from vaultStaticSecret",
		Aliases: []string{"delete"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("helm vault delete", args[0], "-n", namespace)
		},
	}

	getCommon.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")

	upgradeCommon.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")
	upgradeCommon.Flags().StringVarP(&chartValues, "file", "f", "", "File with chart values")
	upgradeCommon.MarkFlagRequired("file")

	deleteCommon.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")

	rootCmd.AddCommand(getCommon)
	rootCmd.AddCommand(upgradeCommon)
	rootCmd.AddCommand(deleteCommon)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func diffArrays(a, b []string) []string {
	var diff []string
	for _, v := range a {
		found := false
		for _, w := range b {
			if v == w {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, v)
		}
	}
	return diff
}

func getUpsertCommonSecrets(configSecrets []string, vaultStaticSecret VaultStaticSecret, deployment string) VaultStaticSecret {
	var commonSecrets []string
	var deploymentList []string

	rolloutRestartTargets := vaultStaticSecret.Spec.RolloutRestartTargets
	for _, vaultSecrets := range rolloutRestartTargets {
		commonSecrets = append(commonSecrets, vaultSecrets.Name)
	}

	diffToCommonSecrets := diffArrays(configSecrets, commonSecrets)

	for _, vaultSecrets := range rolloutRestartTargets {
		deploymentList = append(deploymentList, vaultSecrets.Name)
	}
	diffDeployment := diffArrays([]string{deployment}, deploymentList)

	// fmt.Println(secrets)
	fmt.Println(commonSecrets)
	fmt.Println(diffToCommonSecrets)
	fmt.Println(diffDeployment)

	if len(diffDeployment) > 0 && len(diffToCommonSecrets) > 0 {
		vaultStaticSecret.Spec.RolloutRestartTargets = append(rolloutRestartTargets, struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		}{Kind: "Deployment", Name: deployment})
	}

	return vaultStaticSecret
}

func readYamlFile(filename string) ChartValues {
	yamlFile, err := os.ReadFile(chartValues)
	if err != nil {
		log.Fatalf("Error to read YAML file: %v", err)
	}
	var config ChartValues
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}
	return config
}

func k8sClient() *kubernetes.Clientset {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// Si estamos fuera del clúster de Kubernetes, usar config fuera del cluster
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func getCommonSecrets(clientset *kubernetes.Clientset, namespace string) VaultStaticSecret {

	vaultstaticsecrets, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/secrets.hashicorp.com/v1beta1").
		Resource("vaultstaticsecrets").
		Namespace(namespace).
		Name("common-secrets-v2").
		SetHeader("Accept", "application/yaml").
		DoRaw(context.TODO())

	if err != nil {
		log.Fatalf(err.Error())
	}

	var vaultStaticSecret VaultStaticSecret
	if err := yaml.Unmarshal(vaultstaticsecrets, &vaultStaticSecret); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}
	return vaultStaticSecret
}

package main

import (
	"fmt"
	"helm-plugin-vault/pkg/utils"
	"helm-plugin-vault/pkg/vault"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var namespace string
var filePath string
var vaultEnabled bool
var isDebug bool

var clientset = utils.K8sClient()

func main() {

	var rootCmd = &cobra.Command{Use: "vault"}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	var getCommon = &cobra.Command{
		Use:     "get",
		Short:   "Get common-secrets from vaultStaticSecret",
		Aliases: []string{"get-common"},
		Run: func(cmd *cobra.Command, args []string) {
			if isDebug {
				fmt.Println("namespace:", namespace)
			}
			vaultstaticsecrets := vault.GetVaultCommonSecrets(clientset, namespace, true)
			if vaultstaticsecrets != nil {
				yaml, err := yaml.Marshal(&vaultstaticsecrets)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				fmt.Printf("---\n%s", string(yaml))
			}
		},
	}

	var upgradeCommon = &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade common-secrets from vaultStaticSecret",
		Aliases: []string{"upgrade-common"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if isDebug {
				fmt.Printf("namespace: %s\n\n", namespace)
			}

			chartValues := utils.GetConfigValues(filePath)
			// vaultEnabled = chartValues.VaultSecrets.Enabled

			var vaultEnabled bool
			if _enabled, isEnabled := chartValues["vaultSecrets"].(map[interface{}]interface{})["enabled"]; isEnabled {
				vaultEnabled = _enabled.(bool)
			}

			if isDebug {
				yamlChartValues, _ := yaml.Marshal(&chartValues)
				fmt.Printf("file %s\n---\n%s\n", filePath, string(yamlChartValues))
			}

			if !vaultEnabled {
				fmt.Printf("VaultSecrets is disabled in %s\n", filePath)
				os.Exit(0)
			}

			deployment := args[0]
			vaultStaticSecret := vault.GetVaultCommonSecrets(clientset, namespace, false)
			if vaultStaticSecret == nil {
				vaultStaticSecret = vault.CreateVaultCommonSecret(clientset, namespace, filePath)
			}

			if isDebug {
				yamlVaultStaticSecret, _ := yaml.Marshal(&vaultStaticSecret)
				fmt.Println("Get vaultStaticSecret common-secrets")
				fmt.Printf("---\n%s\n", string(yamlVaultStaticSecret))
			}

			secret := vault.GetK8sCommonSecrets(clientset, namespace)
			if isDebug && len(secret.Kind) > 0 {
				yamlSecret, _ := yaml.Marshal(&secret)
				fmt.Println("Get secret common-secrets")
				fmt.Printf("---\n%s\n", string(yamlSecret))
			}
			rolloutRestartTarget := vault.GetRolloutRestartTarget(filePath, *secret, *vaultStaticSecret, deployment)
			if isDebug {
				yamlRolloutRestartTarget, _ := yaml.Marshal(&rolloutRestartTarget)
				fmt.Printf("RolloutRestartTarget: %s", string(yamlRolloutRestartTarget))
			}
			vault.PatchCommonSecrets(clientset, namespace, rolloutRestartTarget)
		},
	}

	var deleteCommon = &cobra.Command{
		Use:     "delete",
		Short:   "Delete common-secrets from vaultStaticSecret",
		Aliases: []string{"delete-common"},
		Run: func(cmd *cobra.Command, args []string) {
			if isDebug {
				fmt.Println("namespace:", namespace)
			}
			vault.DeleteCommonSecrets(clientset, namespace)
		},
	}

	upgradeCommon.MarkFlagRequired("file")
	upgradeCommon.Flags().StringVarP(&filePath, "file", "f", "", "File with chart values")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")
	rootCmd.PersistentFlags().BoolVarP(&isDebug, "debug", "d", false, "Debug mode")

	helmNamespace := os.Getenv("HELM_NAMESPACE")
	if helmNamespace != "" {
		namespace = helmNamespace
	}
	helmDebug := os.Getenv("HELM_DEBUG")
	if helmDebug == "true" {
		isDebug = true
	}

	rootCmd.AddCommand(getCommon)
	rootCmd.AddCommand(upgradeCommon)
	rootCmd.AddCommand(deleteCommon)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"helm-plugin-vault/pkg/utils"
	"helm-plugin-vault/pkg/vault"
	"os"

	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var namespace string
var filePath string
var vaultEnabled bool

var clientset = utils.K8sClient()

func main() {

	var rootCmd = &cobra.Command{Use: "vault"}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	var getCommon = &cobra.Command{
		Use:     "get",
		Short:   "Get common-secrets from vaultStaticSecret",
		Aliases: []string{"get-common"},
		Run: func(cmd *cobra.Command, args []string) {
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

			chartValues := utils.GetConfigValues(filePath)
			vaultEnabled = chartValues.VaultSecrets.Enabled

			if !vaultEnabled {
				fmt.Printf("VaultSecrets is disabled in %s\n", filePath)
				os.Exit(0)
			}

			deployment := args[0]
			vaultStaticSecret := vault.GetVaultCommonSecrets(clientset, namespace, false)
			if vaultStaticSecret == nil {
				vaultStaticSecret = vault.CreateVaultCommonSecret(clientset, namespace, filePath)
			}

			secret := vault.GetK8sCommonSecrets(clientset, namespace)
			rolloutRestartTarget := vault.GetRolloutRestartTarget(filePath, *secret, *vaultStaticSecret, deployment)
			vault.PatchCommonSecrets(clientset, namespace, rolloutRestartTarget)
		},
	}

	var deleteCommon = &cobra.Command{
		Use:     "delete",
		Short:   "Delete common-secrets from vaultStaticSecret",
		Aliases: []string{"delete-common"},
		Run: func(cmd *cobra.Command, args []string) {
			vault.DeleteCommonSecrets(clientset, namespace)
		},
	}

	getCommon.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")
	upgradeCommon.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")
	upgradeCommon.Flags().StringVarP(&filePath, "file", "f", "", "File with chart values")
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

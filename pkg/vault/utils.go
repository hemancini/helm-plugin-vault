package vault

import (
	"fmt"
	"helm-plugin-vault/pkg/types"
	"helm-plugin-vault/pkg/utils"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func getEnvironment(filePath string) string {
	var environment string
	if strings.Contains(filePath, "dev") || strings.Contains(filePath, "integration") {
		environment = "dev"
	} else if strings.Contains(filePath, "qa") {
		environment = "qa"
	} else if strings.Contains(filePath, "uat") || strings.Contains(filePath, "staging") {
		environment = "uat"
	} else if strings.Contains(filePath, "prod") || strings.Contains(filePath, "production") {
		environment = "prod"
	} else {
		environment = "dev"
	}
	return environment
}

func getUpsertCommonSecrets(chartValues string, secret Secret, vaultStaticSecret VaultStaticSecret, deployment string) VaultStaticSecret {
	var commonSecrets []string
	var deploymentList []string
	rolloutRestartTargets := vaultStaticSecret.Spec.RolloutRestartTargets

	for key := range secret.Data {
		commonSecrets = append(commonSecrets, key)
	}

	for _, vaultSecrets := range rolloutRestartTargets {
		deploymentList = append(deploymentList, vaultSecrets.Name)
	}

	configValues := utils.GetConfigValues(chartValues)
	configSecrets := configValues.Deployment.Secrets

	newConfigSecrets := diffArrays(configSecrets, commonSecrets)
	newConfigCommonSecrets := commonArray(configSecrets, commonSecrets)
	diffDeployment := diffArrays([]string{deployment}, deploymentList)

	// fmt.Println("configSecrets:", configSecrets)
	// fmt.Println("commonSecrets:", commonSecrets)

	// fmt.Println("newConfigSecrets:", newConfigSecrets)
	// fmt.Println("newConfigCommonSecrets:", newConfigCommonSecrets)
	// fmt.Println("diffDeployment:", diffDeployment)

	configValues.Deployment.Secrets = newConfigSecrets
	configValues.Deployment.CommonSecrets = newConfigCommonSecrets

	newYAML, err := yaml.Marshal(&configValues)
	if err != nil {
		log.Fatalf("Error al serializar el YAML: %v", err)
	}

	if err := os.WriteFile(chartValues, newYAML, 0644); err != nil {
		log.Fatalf("Error al escribir el archivo YAML: %v", err)
	} else {
		fmt.Printf("Successfully updated %s\n", chartValues)
	}

	if len(diffDeployment) > 0 && len(newConfigCommonSecrets) > 0 {
		vaultStaticSecret.Spec.RolloutRestartTargets = append(rolloutRestartTargets, struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		}{Kind: "Deployment", Name: deployment})
	}

	return vaultStaticSecret
}

func GetRolloutRestartTarget(chartValues string, secret Secret, vaultStaticSecret VaultStaticSecret, deployment string) []types.RolloutRestartTarget {
	vaultStaticSecretUpdated := getUpsertCommonSecrets(chartValues, secret, vaultStaticSecret, deployment)
	rolloutRestartTargets := vaultStaticSecretUpdated.Spec.RolloutRestartTargets

	var rolloutRestartTarget []types.RolloutRestartTarget
	for _, deployment := range rolloutRestartTargets {
		rolloutRestartTarget = append(rolloutRestartTarget, types.RolloutRestartTarget{
			Kind: "Deployment",
			Name: deployment.Name,
		})
	}
	return rolloutRestartTarget
}

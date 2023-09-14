package utils

import (
	"helm-plugin-vault/pkg/types"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type ChartValues types.ChartValues

func GetConfigValues(filename string) ChartValues {
	var configValues ChartValues

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error to read YAML file: %v", err)
	}
	if err := yaml.Unmarshal(yamlFile, &configValues); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}
	return configValues
}

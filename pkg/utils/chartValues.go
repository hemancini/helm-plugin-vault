package utils

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func GetConfigValues(filename string) map[string]interface{} {
	var configValues map[string]interface{}

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error to read YAML file: %v", err)
	}
	if err := yaml.Unmarshal(yamlFile, &configValues); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}
	return configValues
}

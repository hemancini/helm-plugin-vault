package vault

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
)

func GetK8sCommonSecrets(clientset *kubernetes.Clientset, namespace string) *Secret {
	apiPath := "/api/v1/namespaces/" + namespace + "/secrets/" + resourceCommonName
	data, err := clientset.RESTClient().
		Get().
		AbsPath(apiPath).
		SetHeader("Accept", "application/yaml").
		DoRaw(context.TODO())

	var secret Secret
	if err != nil {
		fmt.Printf("Warning: secret %s not found in namespace %s\n", resourceCommonName, namespace)
		return &secret
	}

	if err := yaml.Unmarshal(data, &secret); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}

	delete(secret.Data, "_raw")
	return &secret
}

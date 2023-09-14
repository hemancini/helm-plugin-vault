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

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	var secret Secret
	if err := yaml.Unmarshal(data, &secret); err != nil {
		log.Fatalf("Error to decode YAML file: %v", err)
	}

	// yaml, err := yaml.Marshal(&secret)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	// fmt.Printf("---\n%s", string(yaml))

	delete(secret.Data, "_raw")
	return &secret
}

package kubernetes

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func New(kubeconfig string) (*kubernetes.Clientset, error) {
	c, err := getConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		apicfg, err := clientcmd.LoadFromFile(kubeconfig)
		if err != nil {
			return nil, err
		}
		cfg := clientcmd.NewDefaultClientConfig(*apicfg, nil)
		log.Fatal(cfg)
		return cfg.ClientConfig()
	}
	return rest.InClusterConfig()
}

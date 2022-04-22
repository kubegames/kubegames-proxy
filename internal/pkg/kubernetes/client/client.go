package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//init k8s client
func InitClient(kubeconfig string) (*kubernetes.Clientset, error) {
	var err error
	var config *rest.Config

	if len(kubeconfig) > 0 {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err == nil {
			return kubernetes.NewForConfig(config)
		}
	}

	if config, err = rest.InClusterConfig(); err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

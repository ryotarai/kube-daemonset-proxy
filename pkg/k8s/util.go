package k8s

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientset() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()

	if err != nil {
		if err != rest.ErrNotInCluster {
			return nil, errors.Wrap(err, "in cluster config")
		}

		// out of cluster
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = "~/.kube/config"
		}
		kubeconfig, err = homedir.Expand(kubeconfig)
		if err != nil {
			return nil, errors.Wrap(err, "homedir expand")
		}

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, errors.Wrap(err, "clientcmd.BuildConfigFromFlags")
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "kubernetes.NewForConfig")
	}

	return clientset, nil
}

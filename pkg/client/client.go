package client

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient contains the Kubernetes clientset and configuration
type K8sClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// NewClient creates a new Kubernetes client from the provided kubeconfig path
func NewClient(kubeconfig string, verbose bool) (*K8sClient, error) {
	// Load kubeconfig
	// First try using in-cluster config if running in a pod
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		// Get the current context from the kubeconfig
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = kubeconfig
		configOverrides := &clientcmd.ConfigOverrides{}

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		clientConfig, err := kubeConfig.ClientConfig()
		if err != nil {
			fmt.Printf("Error building kubeconfig from current context: %v\n", err)
			// Fall back to default config as a last resort
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return nil, fmt.Errorf("error building default kubeconfig: %v", err)
			}
		} else {
			config = clientConfig
		}
	}

	if verbose {
		fmt.Printf("Using Kubernetes API at: %s\n", config.Host)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	return &K8sClient{
		Clientset: clientset,
		Config:    config,
	}, nil
}
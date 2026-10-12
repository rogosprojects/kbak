package client

import (
	"fmt"

	"github.com/rogosprojects/kbak/pkg/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient contains the Kubernetes clientset and configuration
type K8sClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	// Namespace is the default namespace: the current context's, or the pod's when running in-cluster
	Namespace string
}

// NewClient creates a Kubernetes client the same way kubectl does: the explicit kubeconfig path if set,
// otherwise $KUBECONFIG or ~/.kube/config, otherwise the in-cluster service account.
func NewClient(kubeconfig string, verbose bool) (*K8sClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading Kubernetes configuration: %v", err)
	}

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, fmt.Errorf("error getting current namespace: %v", err)
	}

	if verbose {
		fmt.Printf("%s %s%sUsing Kubernetes API at: %s%s\n",
			utils.K8sEmoji, utils.Blue, utils.Bold, config.Host, utils.Reset)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	return &K8sClient{
		Clientset: clientset,
		Config:    config,
		Namespace: namespace,
	}, nil
}

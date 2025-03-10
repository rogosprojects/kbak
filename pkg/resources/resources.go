package resources

import (
	"context"
	"strings"

	"github.com/rogosprojects/kbak/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceType defines a Kubernetes resource type that can be backed up
type ResourceType struct {
	Kind    string
	APIFunc func(client *client.K8sClient, namespace string, opts metav1.ListOptions) (interface{}, error)
}

// GetAllResourceTypes returns all supported Kubernetes resource types
func GetAllResourceTypes() []ResourceType {
	return []ResourceType{
		{
			Kind: "Pod",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().Pods(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Deployment",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.AppsV1().Deployments(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Service",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().Services(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "ConfigMap",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().ConfigMaps(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Secret",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().Secrets(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "PersistentVolumeClaim",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().PersistentVolumeClaims(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "ServiceAccount",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.CoreV1().ServiceAccounts(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "StatefulSet",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.AppsV1().StatefulSets(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "DaemonSet",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.AppsV1().DaemonSets(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Ingress",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.NetworkingV1().Ingresses(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Role",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.RbacV1().Roles(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "RoleBinding",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.RbacV1().RoleBindings(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "CronJob",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.BatchV1().CronJobs(ns).List(context.TODO(), opts)
			},
		},
		{
			Kind: "Job",
			APIFunc: func(client *client.K8sClient, ns string, opts metav1.ListOptions) (interface{}, error) {
				return client.Clientset.BatchV1().Jobs(ns).List(context.TODO(), opts)
			},
		},
	}
}

// GetResourceTypes returns a filtered list of Kubernetes resource types to backup
// If selectedTypes is empty, all resource types are returned
func GetResourceTypes(selectedTypes map[string]bool) []ResourceType {
	allTypes := GetAllResourceTypes()

	// If no resource types are specified, return all types
	if len(selectedTypes) == 0 {
		return allTypes
	}

	// Filter resource types based on selection
	var filteredTypes []ResourceType
	for _, resourceType := range allTypes {
		if selectedTypes[strings.ToLower(resourceType.Kind)] {
			filteredTypes = append(filteredTypes, resourceType)
		}
	}

	return filteredTypes
}

// IsNotFoundError checks if an error is a "not found" error
func IsNotFoundError(err error) bool {
	// Handle nil error case
	if err == nil {
		return false
	}

	// Check if the error message contains common "not found" indicators
	errMsg := err.Error()
	notFoundIndicators := []string{
		"not found",
		"the server could not find the requested resource",
		"no matches for kind",
		"the server doesn't have a resource type",
	}

	for _, indicator := range notFoundIndicators {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

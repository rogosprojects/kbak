package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

func main() {
	var namespace string
	var kubeconfig string
	var outputDir string
	var verbose bool

	flag.StringVar(&namespace, "namespace", "", "Namespace to backup (required)")
	flag.StringVar(&outputDir, "output", "backup", "Output directory for backup files")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	}

	flag.Parse()

	if namespace == "" {
		fmt.Println("Error: --namespace flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create output directory with timestamp
	timestamp := time.Now().Format("02Jan2006-15:04")
	backupDir := filepath.Join(outputDir, timestamp, namespace)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting backup of namespace '%s' to '%s'\n", namespace, backupDir)

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
				fmt.Printf("Error building default kubeconfig: %v\n", err)
				os.Exit(1)
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
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Dynamic client has been removed as we don't need it without custom resources

	// Resources to backup
	resourceTypes := []struct {
		kind    string
		apiFunc func(string, metav1.ListOptions) (interface{}, error)
	}{
		{
			kind: "Pod",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().Pods(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Deployment",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.AppsV1().Deployments(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Service",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().Services(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "ConfigMap",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().ConfigMaps(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Secret",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().Secrets(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "PersistentVolumeClaim",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().PersistentVolumeClaims(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "ServiceAccount",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.CoreV1().ServiceAccounts(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "StatefulSet",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.AppsV1().StatefulSets(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "DaemonSet",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.AppsV1().DaemonSets(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Ingress",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.NetworkingV1().Ingresses(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Role",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.RbacV1().Roles(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "RoleBinding",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.RbacV1().RoleBindings(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "CronJob",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.BatchV1().CronJobs(ns).List(context.TODO(), opts)
			},
		},
		{
			kind: "Job",
			apiFunc: func(ns string, opts metav1.ListOptions) (interface{}, error) {
				return clientset.BatchV1().Jobs(ns).List(context.TODO(), opts)
			},
		},
	}

	// Custom resources have been removed

	// Backup each resource type
	resourceCount := 0
	errorCount := 0

	for _, resource := range resourceTypes {
		var objects interface{}
		var err error

		// Get resources from the Kubernetes API
		objects, err = resource.apiFunc(namespace, metav1.ListOptions{})

		if err != nil {
			fmt.Printf("Error listing %s: %v\n", resource.kind, err)
			if verbose {
				fmt.Printf("Debug info - API endpoint: %s\n", config.Host)
				fmt.Printf("Debug info - Resource: %s in namespace %s\n", resource.kind, namespace)
			}
			errorCount++
			continue
		}

		// Debug the response from the API
		if verbose {
			fmt.Printf("Response type for %s: %T\n", resource.kind, objects)
		}

		// Extract items from the list
		items, itemCount := extractItems(objects)
		if verbose {
			fmt.Printf("Found %d %s resources in namespace %s\n", itemCount, resource.kind, namespace)
		}
		if itemCount == 0 {
			// Skip creating directories for resource kinds with no items
			continue
		}

		// Create directory for this resource kind
		kindDir := filepath.Join(backupDir, resource.kind)
		if err := os.MkdirAll(kindDir, 0755); err != nil {
			fmt.Printf("Error creating directory for %s: %v\n", resource.kind, err)
			errorCount++
			continue
		}

		itemsBackedUp := 0
		for i, item := range items {
			if item == nil {
				continue
			}

			name := extractName(item)
			if name == "" {
				name = fmt.Sprintf("unknown-%d", i)
			}

			// Remove cluster-specific and runtime fields
			cleanObject(item)

			// Convert to YAML
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				fmt.Printf("Error marshaling %s '%s': %v\n", resource.kind, name, err)
				continue
			}

			// Save to file
			filename := filepath.Join(kindDir, name+".yaml")
			if err := os.WriteFile(filename, yamlData, 0644); err != nil {
				fmt.Printf("Error writing %s '%s': %v\n", resource.kind, name, err)
				continue
			}

			itemsBackedUp++
		}

		if itemsBackedUp > 0 {
			fmt.Printf("Backed up %d %s resources\n", itemsBackedUp, resource.kind)
			resourceCount += itemsBackedUp
		}
	}

	if errorCount > 0 {
		fmt.Printf("Completed with %d errors\n", errorCount)
	}

	if resourceCount > 0 {
		fmt.Printf("Backup completed successfully to %s (%d resources total)\n", backupDir, resourceCount)
	} else {
		fmt.Printf("No resources found to backup in namespace '%s'\n", namespace)
	}
}

// extractItems gets the items slice from various list types
func extractItems(list interface{}) ([]interface{}, int) {
	// Check for different list types and extract items accordingly
	switch v := list.(type) {
	case metav1.ListInterface:
		items := extractItemsFromListInterface(v)
		remainingCount := 0
		if v.GetRemainingItemCount() != nil {
			remainingCount = int(*v.GetRemainingItemCount())
		}
		return items, len(items) + remainingCount
	default:
		// Use reflection as a fallback for types we don't directly support
		items, count := extractItemsUsingReflection(v)
		return items, count
	}
}

// extractItemsFromListInterface extracts items from a kubernetes ListInterface
func extractItemsFromListInterface(list metav1.ListInterface) []interface{} {
	// Type assertions for common list types
	// This would need to be expanded for all resource types
	switch v := list.(type) {
	case *metav1.List:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			result[i] = item
		}
		return result
	default:
		// Fallback using reflection
		items, _ := extractItemsUsingReflection(list)
		return items
	}
}

// extractItemsUsingReflection attempts to extract items using type assertions
func extractItemsUsingReflection(obj interface{}) ([]interface{}, int) {
	// For specific kubernetes types
	switch v := obj.(type) {
	case *v1.PodList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item // Create a new variable to avoid issues with loop variable capture
			result[i] = &item
		}
		return result, len(result)

	case *appsv1.DeploymentList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item
			result[i] = &item
		}
		return result, len(result)

	case *v1.ServiceList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item
			result[i] = &item
		}
		return result, len(result)

	// For all other types, use a more generic approach
	default:
		// Try to handle list types via type assertions
		// For unstructured lists
		if list, ok := obj.(map[string]interface{}); ok {
			if items, ok := list["items"].([]interface{}); ok {
				return items, len(items)
			}
		}

		// Try to access via methods if available
		if list, ok := obj.(interface{ GetItems() []interface{} }); ok {
			items := list.GetItems()
			return items, len(items)
		}

		// Try other common variants that might be used
		if list, ok := obj.(interface{ Items() []interface{} }); ok {
			return list.Items(), len(list.Items())
		}

		// Return an empty list for unhandled types
		//fmt.Printf("Warning: Unhandled type in extractItemsUsingReflection: %T\n", obj)
		return []interface{}{}, 0
	}
}

// extractName attempts to get the name of a Kubernetes resource
// isNotFoundError checks if an error is a "not found" error
func isNotFoundError(err error) bool {
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

func extractName(obj interface{}) string {
	// Try to access metadata.name
	if metaObj, ok := obj.(metav1.Object); ok {
		return metaObj.GetName()
	}

	// For unstructured objects from dynamic client
	if unstr, ok := obj.(map[string]interface{}); ok {
		if metadata, ok := unstr["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				return name
			}
		}
	}

	return ""
}

// cleanObject removes fields that shouldn't be included in backups
func cleanObject(obj interface{}) {
	// Handle different types of Kubernetes objects
	switch typedObj := obj.(type) {
	// Pod
	case *v1.Pod:
		cleanPod(typedObj)

	// Deployment
	case *appsv1.Deployment:
		cleanDeployment(typedObj)

	// Service
	case *v1.Service:
		cleanService(typedObj)

	// For all other types, try as unstructured
	default:
		// For unstructured objects
		if unstr, ok := obj.(map[string]interface{}); ok {
			// Remove status
			delete(unstr, "status")

			// Clean metadata
			if metadata, ok := unstr["metadata"].(map[string]interface{}); ok {
				fieldsToDelete := []string{
					"creationTimestamp",
					"resourceVersion",
					"selfLink",
					"uid",
					"generation",
					"managedFields",
					"annotations",
					"ownerReferences",
				}

				for _, field := range fieldsToDelete {
					delete(metadata, field)
				}

				// Remove system annotations but keep user ones
				if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
					for k := range annotations {
						if isSystemAnnotation(k) {
							delete(annotations, k)
						}
					}
					// If no annotations left, remove the map
					if len(annotations) == 0 {
						delete(metadata, "annotations")
					}
				}
			}
		}
	}
}

// isSystemAnnotation checks if an annotation is system-managed
func isSystemAnnotation(key string) bool {
	systemPrefixes := []string{
		"kubernetes.io/",
		"k8s.io/",
		"control-plane.alpha.kubernetes.io/",
		"app.kubernetes.io/",
		"cni.projectcalico.org/",
		"kubectl.kubernetes.io/",
		"deployment.kubernetes.io/",
		"meta.helm.sh/",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

// cleanPod removes server-side fields from a Pod
func cleanPod(pod *v1.Pod) {
	// Remove status
	pod.Status = v1.PodStatus{}

	// Clean metadata but preserve essential fields
	cleanMetadata(&pod.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	pod.APIVersion = "v1"
	pod.Kind = "Pod"

	// Remove generated fields but keep essential ones
	pod.Spec.NodeName = ""
	pod.Spec.DeprecatedServiceAccount = ""

	// Retain the service account name if it's not "default"
	if pod.Spec.ServiceAccountName == "default" {
		pod.Spec.ServiceAccountName = ""
	}
}

// cleanDeployment removes server-side fields from a Deployment
func cleanDeployment(deployment *appsv1.Deployment) {
	// Remove status
	deployment.Status = appsv1.DeploymentStatus{}

	// Clean metadata
	cleanMetadata(&deployment.ObjectMeta)

	// Clean template metadata but preserve essential fields
	cleanMetadata(&deployment.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	deployment.APIVersion = "apps/v1"
	deployment.Kind = "Deployment"

	// Note: Pod templates don't have APIVersion field

	// Ensure selector is preserved (it's required for deployments)
	if deployment.Spec.Selector == nil || len(deployment.Spec.Selector.MatchLabels) == 0 {
		// If no selector, create one that matches template labels
		if deployment.Spec.Template.Labels != nil && len(deployment.Spec.Template.Labels) > 0 {
			if deployment.Spec.Selector == nil {
				deployment.Spec.Selector = &metav1.LabelSelector{}
			}
			deployment.Spec.Selector.MatchLabels = deployment.Spec.Template.Labels
		}
	}
}

// cleanService removes server-side fields from a Service
func cleanService(svc *v1.Service) {
	// Remove status
	svc.Status = v1.ServiceStatus{}

	// Clean metadata
	cleanMetadata(&svc.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	svc.APIVersion = "v1"
	svc.Kind = "Service"

	// Handle ClusterIP specially
	if svc.Spec.ClusterIP == "None" {
		// Preserve "None" for headless services
		svc.Spec.ClusterIP = "None"
	} else {
		// Remove dynamic ClusterIP
		svc.Spec.ClusterIP = ""
	}

	// Reset other fields that are server-populated
	svc.Spec.ClusterIPs = nil
	svc.Spec.ExternalIPs = nil
	svc.Spec.LoadBalancerIP = ""
	svc.Spec.LoadBalancerSourceRanges = nil

	// Preserve intentionally set fields
	// Only clear these if they're the default values
	if svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeCluster {
		svc.Spec.ExternalTrafficPolicy = ""
	}

	if svc.Spec.SessionAffinity == v1.ServiceAffinityNone {
		svc.Spec.SessionAffinity = ""
		svc.Spec.SessionAffinityConfig = nil
	}

	// Keep PublishNotReadyAddresses if it's true (intentionally set)
	if !svc.Spec.PublishNotReadyAddresses {
		svc.Spec.PublishNotReadyAddresses = false
	}

	// Always clear health check node port as it's assigned by the system
	svc.Spec.HealthCheckNodePort = 0
}

// cleanMetadata removes server-side fields from ObjectMeta
func cleanMetadata(meta *metav1.ObjectMeta) {
	// Preserve namespace and name
	name := meta.Name
	namespace := meta.Namespace
	labels := meta.Labels
	annotations := meta.Annotations

	// Remove timestamps, versions, and UIDs
	meta.CreationTimestamp = metav1.Time{}
	meta.DeletionTimestamp = nil
	meta.DeletionGracePeriodSeconds = nil
	meta.Generation = 0
	meta.ResourceVersion = ""
	meta.SelfLink = ""
	meta.UID = ""
	meta.ManagedFields = nil

	// Remove owner references as they're server-side
	meta.OwnerReferences = nil

	// Remove finalizers
	meta.Finalizers = nil

	// Restore essential fields
	meta.Name = name
	meta.Namespace = namespace
	meta.Labels = labels

	// Clean annotations but keep user ones
	if annotations != nil {
		for k := range annotations {
			if isSystemAnnotation(k) {
				delete(annotations, k)
			}
		}
		if len(annotations) == 0 {
			meta.Annotations = nil
		} else {
			meta.Annotations = annotations
		}
	}
}

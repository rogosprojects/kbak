package utils

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CleanObject removes fields that shouldn't be included in backups
func CleanObject(obj interface{}) {
	// Handle different types of Kubernetes objects
	switch typedObj := obj.(type) {
	// Pod
	case *v1.Pod:
		CleanPod(typedObj)

	// Deployment
	case *appsv1.Deployment:
		CleanDeployment(typedObj)

	// Service
	case *v1.Service:
		CleanService(typedObj)

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
						if IsSystemAnnotation(k) {
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

// IsSystemAnnotation checks if an annotation is system-managed
func IsSystemAnnotation(key string) bool {
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

// CleanPod removes server-side fields from a Pod
func CleanPod(pod *v1.Pod) {
	// Remove status
	pod.Status = v1.PodStatus{}

	// Clean metadata but preserve essential fields
	CleanMetadata(&pod.ObjectMeta)

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

// CleanDeployment removes server-side fields from a Deployment
func CleanDeployment(deployment *appsv1.Deployment) {
	// Remove status
	deployment.Status = appsv1.DeploymentStatus{}

	// Clean metadata
	CleanMetadata(&deployment.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&deployment.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	deployment.APIVersion = "apps/v1"
	deployment.Kind = "Deployment"

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

// CleanService removes server-side fields from a Service
func CleanService(svc *v1.Service) {
	// Remove status
	svc.Status = v1.ServiceStatus{}

	// Clean metadata
	CleanMetadata(&svc.ObjectMeta)

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

// CleanMetadata removes server-side fields from ObjectMeta
func CleanMetadata(meta *metav1.ObjectMeta) {
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
			if IsSystemAnnotation(k) {
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
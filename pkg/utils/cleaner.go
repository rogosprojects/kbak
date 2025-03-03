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
		
	// ConfigMap
	case *v1.ConfigMap:
		CleanConfigMap(typedObj)
		
	// Secret
	case *v1.Secret:
		CleanSecret(typedObj)
		
	// PersistentVolumeClaim
	case *v1.PersistentVolumeClaim:
		CleanPVC(typedObj)
		
	// ServiceAccount
	case *v1.ServiceAccount:
		CleanServiceAccount(typedObj)

	// For all other types, try as unstructured
	default:
		// For unstructured objects
		if unstr, ok := obj.(map[string]interface{}); ok {
			// Remove status
			delete(unstr, "status")

			// Make sure apiVersion and kind are preserved
			// Do not delete these fields under any circumstances
			apiVersion, apiVersionExists := unstr["apiVersion"]
			kind, kindExists := unstr["kind"]
			
			// If either field is missing or empty, try to infer them
			if !apiVersionExists || apiVersion == "" || !kindExists || kind == "" {
				inferAPIVersionAndKind(unstr)
			}

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

// inferAPIVersionAndKind attempts to set apiVersion and kind for unstructured objects
// if they're missing. This is a fallback mechanism.
func inferAPIVersionAndKind(obj map[string]interface{}) {
	// Try to infer apiVersion if kind exists
	if kind, ok := obj["kind"].(string); ok && kind != "" {
		// Set apiVersion based on kind if possible
		switch kind {
		case "Pod", "Service", "ConfigMap", "Secret", "ServiceAccount", "PersistentVolumeClaim", "Namespace":
			obj["apiVersion"] = "v1"
		case "Deployment", "StatefulSet", "DaemonSet", "ReplicaSet":
			obj["apiVersion"] = "apps/v1"
		case "Ingress", "IngressClass", "NetworkPolicy":
			obj["apiVersion"] = "networking.k8s.io/v1"
		case "Role", "RoleBinding", "ClusterRole", "ClusterRoleBinding":
			obj["apiVersion"] = "rbac.authorization.k8s.io/v1"
		case "Job", "CronJob":
			obj["apiVersion"] = "batch/v1"
		case "HorizontalPodAutoscaler":
			obj["apiVersion"] = "autoscaling/v2"
		case "PodDisruptionBudget":
			obj["apiVersion"] = "policy/v1"
		case "CustomResourceDefinition":
			obj["apiVersion"] = "apiextensions.k8s.io/v1"
		}
	}
	
	// Try to infer kind if apiVersion exists
	if apiVersion, ok := obj["apiVersion"].(string); ok && apiVersion != "" {
		// If there's already a kind, don't overwrite it
		if _, hasKind := obj["kind"].(string); !hasKind {
			switch apiVersion {
			case "v1":
				// Default to Pod if we have to guess
				obj["kind"] = "Pod"
			case "apps/v1":
				// Default to Deployment
				obj["kind"] = "Deployment"
			case "batch/v1":
				// Default to Job
				obj["kind"] = "Job"
			case "networking.k8s.io/v1":
				// Default to Ingress
				obj["kind"] = "Ingress"
			case "rbac.authorization.k8s.io/v1":
				// Default to Role
				obj["kind"] = "Role"
			}
		}
	}
	
	// If still no apiVersion and kind, try to infer from metadata or resource structure
	if _, hasApiVersion := obj["apiVersion"].(string); !hasApiVersion {
		if _, hasKind := obj["kind"].(string); !hasKind {
			// Try to infer from metadata.name patterns
			if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok && name != "" {
					// Name-based heuristic
					nameLower := strings.ToLower(name)
					if strings.Contains(nameLower, "deploy") {
						obj["kind"] = "Deployment"
						obj["apiVersion"] = "apps/v1"
					} else if strings.Contains(nameLower, "svc") || strings.Contains(nameLower, "service") {
						obj["kind"] = "Service"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "cm") || strings.Contains(nameLower, "config") {
						obj["kind"] = "ConfigMap"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "secret") {
						obj["kind"] = "Secret"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "pod") {
						obj["kind"] = "Pod"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "job") {
						obj["kind"] = "Job"
						obj["apiVersion"] = "batch/v1"
					} else if strings.Contains(nameLower, "cron") {
						obj["kind"] = "CronJob"
						obj["apiVersion"] = "batch/v1"
					} else if strings.Contains(nameLower, "ing") {
						obj["kind"] = "Ingress"
						obj["apiVersion"] = "networking.k8s.io/v1"
					} else if strings.Contains(nameLower, "role") && !strings.Contains(nameLower, "cluster") {
						obj["kind"] = "Role"
						obj["apiVersion"] = "rbac.authorization.k8s.io/v1"
					} else if strings.Contains(nameLower, "binding") && !strings.Contains(nameLower, "cluster") {
						obj["kind"] = "RoleBinding"
						obj["apiVersion"] = "rbac.authorization.k8s.io/v1"
					} else if strings.Contains(nameLower, "sa") || strings.Contains(nameLower, "serviceaccount") {
						obj["kind"] = "ServiceAccount"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "pvc") || strings.Contains(nameLower, "claim") {
						obj["kind"] = "PersistentVolumeClaim"
						obj["apiVersion"] = "v1"
					} else if strings.Contains(nameLower, "sts") || strings.Contains(nameLower, "stateful") {
						obj["kind"] = "StatefulSet"
						obj["apiVersion"] = "apps/v1"
					} else if strings.Contains(nameLower, "ds") || strings.Contains(nameLower, "daemon") {
						obj["kind"] = "DaemonSet"
						obj["apiVersion"] = "apps/v1"
					}
				}
			}
			
			// If still no values, set defaults as last resort
			if _, hasApiVersion := obj["apiVersion"].(string); !hasApiVersion {
				obj["apiVersion"] = "v1"
			}
			if _, hasKind := obj["kind"].(string); !hasKind {
				obj["kind"] = "Pod" // Default to Pod as last resort
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

// CleanConfigMap removes server-side fields from a ConfigMap
func CleanConfigMap(cm *v1.ConfigMap) {
	// Clean metadata
	CleanMetadata(&cm.ObjectMeta)
	
	// Set API version and kind for valid Kubernetes manifests
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	
	// Keep the data and binaryData fields intact
	// No need to modify the actual ConfigMap data
}

// CleanSecret removes server-side fields from a Secret
func CleanSecret(secret *v1.Secret) {
	// Clean metadata
	CleanMetadata(&secret.ObjectMeta)
	
	// Set API version and kind for valid Kubernetes manifests
	secret.APIVersion = "v1"
	secret.Kind = "Secret"
	
	// Keep the data and type fields intact
	// The actual Secret data should be preserved
}

// CleanPVC removes server-side fields from a PersistentVolumeClaim
func CleanPVC(pvc *v1.PersistentVolumeClaim) {
	// Clean metadata
	CleanMetadata(&pvc.ObjectMeta)
	
	// Set API version and kind for valid Kubernetes manifests
	pvc.APIVersion = "v1"
	pvc.Kind = "PersistentVolumeClaim"
	
	// Remove status
	pvc.Status = v1.PersistentVolumeClaimStatus{}
	
	// Clean spec but keep essential fields
	// Don't modify storageClassName, accessModes, resources
}

// CleanServiceAccount removes server-side fields from a ServiceAccount
func CleanServiceAccount(sa *v1.ServiceAccount) {
	// Clean metadata
	CleanMetadata(&sa.ObjectMeta)
	
	// Set API version and kind for valid Kubernetes manifests
	sa.APIVersion = "v1"
	sa.Kind = "ServiceAccount"
	
	// Clean other server-generated fields
	sa.Secrets = nil        // Server manages the secrets references
	sa.ImagePullSecrets = nil  // Only keep manually added ones
	
	// Keep essential fields like automountServiceAccountToken if set
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
package utils

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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

	// StatefulSet
	case *appsv1.StatefulSet:
		CleanStatefulSet(typedObj)

	// DaemonSet
	case *appsv1.DaemonSet:
		CleanDaemonSet(typedObj)

	// ReplicaSet
	case *appsv1.ReplicaSet:
		CleanReplicaSet(typedObj)

	// Job
	case *batchv1.Job:
		CleanJob(typedObj)

	// CronJob
	case *batchv1.CronJob:
		CleanCronJob(typedObj)

	// Ingress
	case *networkingv1.Ingress:
		CleanIngress(typedObj)

	// PodDisruptionBudget
	case *policyv1.PodDisruptionBudget:
		CleanPDB(typedObj)

	// Role and ClusterRole
	case *rbacv1.Role:
		CleanRole(typedObj)
	case *rbacv1.ClusterRole:
		CleanClusterRole(typedObj)

	// RoleBinding and ClusterRoleBinding
	case *rbacv1.RoleBinding:
		CleanRoleBinding(typedObj)
	case *rbacv1.ClusterRoleBinding:
		CleanClusterRoleBinding(typedObj)

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

// IsSystemLabel checks if a label is system-managed for Jobs
func IsSystemLabel(key string) bool {
	systemLabelPrefixes := []string{
		"batch.kubernetes.io/",
	}

	systemLabels := []string{
		"controller-uid",
	}

	// Check for system label prefixes
	for _, prefix := range systemLabelPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Check for specific system labels
	for _, label := range systemLabels {
		if key == label {
			return true
		}
	}

	return false
}

// CleanJobLabels removes system-managed labels from Job metadata
func CleanJobLabels(meta *metav1.ObjectMeta) {
	if meta.Labels != nil {
		cleanedLabels := make(map[string]string)
		for k, v := range meta.Labels {
			if !IsSystemLabel(k) {
				cleanedLabels[k] = v
			}
		}
		if len(cleanedLabels) == 0 {
			meta.Labels = nil
		} else {
			meta.Labels = cleanedLabels
		}
	}
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
	sa.Secrets = nil          // Server manages the secrets references
	sa.ImagePullSecrets = nil // Only keep manually added ones

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
	meta.UID = ""
	meta.SelfLink = ""
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
		cleanedAnnotations := make(map[string]string)
		for k, v := range annotations {
			if !IsSystemAnnotation(k) {
				cleanedAnnotations[k] = v
			}
		}
		if len(cleanedAnnotations) == 0 {
			meta.Annotations = nil
		} else {
			meta.Annotations = cleanedAnnotations
		}
	} else {
		meta.Annotations = nil
	}
}

// CleanStatefulSet removes server-side fields from a StatefulSet
func CleanStatefulSet(statefulset *appsv1.StatefulSet) {
	// Remove status
	statefulset.Status = appsv1.StatefulSetStatus{}

	// Clean metadata
	CleanMetadata(&statefulset.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&statefulset.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	statefulset.APIVersion = "apps/v1"
	statefulset.Kind = "StatefulSet"

	// Ensure selector is preserved (it's required for statefulsets)
	if statefulset.Spec.Selector == nil || len(statefulset.Spec.Selector.MatchLabels) == 0 {
		// If no selector, create one that matches template labels
		if statefulset.Spec.Template.Labels != nil && len(statefulset.Spec.Template.Labels) > 0 {
			if statefulset.Spec.Selector == nil {
				statefulset.Spec.Selector = &metav1.LabelSelector{}
			}
			statefulset.Spec.Selector.MatchLabels = statefulset.Spec.Template.Labels
		}
	}
}

// CleanDaemonSet removes server-side fields from a DaemonSet
func CleanDaemonSet(daemonset *appsv1.DaemonSet) {
	// Remove status
	daemonset.Status = appsv1.DaemonSetStatus{}

	// Clean metadata
	CleanMetadata(&daemonset.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&daemonset.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	daemonset.APIVersion = "apps/v1"
	daemonset.Kind = "DaemonSet"

	// Ensure selector is preserved (it's required for daemonsets)
	if daemonset.Spec.Selector == nil || len(daemonset.Spec.Selector.MatchLabels) == 0 {
		// If no selector, create one that matches template labels
		if daemonset.Spec.Template.Labels != nil && len(daemonset.Spec.Template.Labels) > 0 {
			if daemonset.Spec.Selector == nil {
				daemonset.Spec.Selector = &metav1.LabelSelector{}
			}
			daemonset.Spec.Selector.MatchLabels = daemonset.Spec.Template.Labels
		}
	}
}

// CleanReplicaSet removes server-side fields from a ReplicaSet
func CleanReplicaSet(replicaset *appsv1.ReplicaSet) {
	// Remove status
	replicaset.Status = appsv1.ReplicaSetStatus{}

	// Clean metadata
	CleanMetadata(&replicaset.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&replicaset.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	replicaset.APIVersion = "apps/v1"
	replicaset.Kind = "ReplicaSet"

	// Ensure selector is preserved (it's required for replicasets)
	if replicaset.Spec.Selector == nil || len(replicaset.Spec.Selector.MatchLabels) == 0 {
		// If no selector, create one that matches template labels
		if replicaset.Spec.Template.Labels != nil && len(replicaset.Spec.Template.Labels) > 0 {
			if replicaset.Spec.Selector == nil {
				replicaset.Spec.Selector = &metav1.LabelSelector{}
			}
			replicaset.Spec.Selector.MatchLabels = replicaset.Spec.Template.Labels
		}
	}
}

// CleanJob removes server-side fields from a Job
func CleanJob(job *batchv1.Job) {
	// Remove status
	job.Status = batchv1.JobStatus{}

	// Clean metadata
	CleanMetadata(&job.ObjectMeta)

	// Clean Job-specific system labels (controller-uid, batch.kubernetes.io/*)
	CleanJobLabels(&job.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&job.Spec.Template.ObjectMeta)

	// Also clean Job-specific labels from the template metadata
	CleanJobLabels(&job.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	job.APIVersion = "batch/v1"
	job.Kind = "Job"

	// Reset fields that are typically server-assigned
	job.Spec.Selector = nil // Selector is automatically generated for Jobs

	// Clean up Job-specific server-managed fields
	// These fields can accumulate and cause memory issues in large Job lists
	if job.Spec.ManualSelector != nil && !*job.Spec.ManualSelector {
		job.Spec.ManualSelector = nil
	}

	// Reset completion and failure tracking fields that are server-managed
	// Keep user-specified values for parallelism, completions, etc.
	// but remove server-assigned tracking fields
}

// CleanCronJob removes server-side fields from a CronJob
func CleanCronJob(cronjob *batchv1.CronJob) {
	// Remove status
	cronjob.Status = batchv1.CronJobStatus{}

	// Clean metadata
	CleanMetadata(&cronjob.ObjectMeta)

	// Clean template metadata but preserve essential fields
	CleanMetadata(&cronjob.Spec.JobTemplate.ObjectMeta)
	CleanMetadata(&cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta)

	// Clean Job-specific system labels from the job template
	CleanJobLabels(&cronjob.Spec.JobTemplate.ObjectMeta)
	CleanJobLabels(&cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	cronjob.APIVersion = "batch/v1"
	cronjob.Kind = "CronJob"

	// Reset fields that are typically server-assigned
	cronjob.Spec.JobTemplate.Spec.Selector = nil // Selector is automatically generated for Jobs
}

// CleanIngress removes server-side fields from an Ingress
func CleanIngress(ingress *networkingv1.Ingress) {
	// Remove status
	ingress.Status = networkingv1.IngressStatus{}

	// Clean metadata
	CleanMetadata(&ingress.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	ingress.APIVersion = "networking.k8s.io/v1"
	ingress.Kind = "Ingress"
}

// CleanPDB removes server-side fields from a PodDisruptionBudget
func CleanPDB(pdb *policyv1.PodDisruptionBudget) {
	// Remove status
	pdb.Status = policyv1.PodDisruptionBudgetStatus{}

	// Clean metadata
	CleanMetadata(&pdb.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	pdb.APIVersion = "policy/v1"
	pdb.Kind = "PodDisruptionBudget"
}

// CleanRole removes server-side fields from a Role
func CleanRole(role *rbacv1.Role) {
	// Clean metadata
	CleanMetadata(&role.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	role.APIVersion = "rbac.authorization.k8s.io/v1"
	role.Kind = "Role"
}

// CleanClusterRole removes server-side fields from a ClusterRole
func CleanClusterRole(clusterRole *rbacv1.ClusterRole) {
	// Clean metadata
	CleanMetadata(&clusterRole.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	clusterRole.APIVersion = "rbac.authorization.k8s.io/v1"
	clusterRole.Kind = "ClusterRole"
}

// CleanRoleBinding removes server-side fields from a RoleBinding
func CleanRoleBinding(roleBinding *rbacv1.RoleBinding) {
	// Clean metadata
	CleanMetadata(&roleBinding.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	roleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	roleBinding.Kind = "RoleBinding"
}

// CleanClusterRoleBinding removes server-side fields from a ClusterRoleBinding
func CleanClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) {
	// Clean metadata
	CleanMetadata(&clusterRoleBinding.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	clusterRoleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	clusterRoleBinding.Kind = "ClusterRoleBinding"
}

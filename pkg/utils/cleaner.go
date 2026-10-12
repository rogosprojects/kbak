package utils

import (
	"slices"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

	// Job
	case *batchv1.Job:
		CleanJob(typedObj)

	// CronJob
	case *batchv1.CronJob:
		CleanCronJob(typedObj)

	// Ingress
	case *networkingv1.Ingress:
		CleanIngress(typedObj)

	// Role
	case *rbacv1.Role:
		CleanRole(typedObj)

	// RoleBinding
	case *rbacv1.RoleBinding:
		CleanRoleBinding(typedObj)

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

	// Route (OpenShift)
	case *routev1.Route:
		CleanRoute(typedObj)
	}
}

// IsSystemAnnotation checks if an annotation is system-managed
func IsSystemAnnotation(key string) bool {
	switch key {
	// User-set keys under system prefixes that a restore needs
	case "kubernetes.io/ingress.class", "kubernetes.io/service-account.name":
		return false
	case "volume.beta.kubernetes.io/storage-provisioner",
		// OpenShift: admitting SCC, and the generated pull secret it re-creates for a restored ServiceAccount
		"openshift.io/scc", "openshift.io/internal-registry-pull-secret-ref",
		// Multus network status (the user-set request is k8s.v1.cni.cncf.io/networks and is kept)
		"k8s.v1.cni.cncf.io/network-status", "k8s.v1.cni.cncf.io/networks-status":
		return true
	}

	systemPrefixes := []string{
		"kubernetes.io/",
		"k8s.io/",
		"control-plane.alpha.kubernetes.io/",
		"app.kubernetes.io/",
		"cni.projectcalico.org/",
		"kubectl.kubernetes.io/",
		"deployment.kubernetes.io/",
		"meta.helm.sh/",
		"pv.kubernetes.io/",
		"volume.kubernetes.io/",
		"k8s.ovn.org/",
		// Status the OpenShift service CA writes on a Service; serving-cert-secret-name is the user's request
		"service.beta.openshift.io/serving-cert-signed-by",
		"service.alpha.openshift.io/serving-cert-signed-by",
		"service.beta.openshift.io/serving-cert-generation-error",
		"service.alpha.openshift.io/serving-cert-generation-error",
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
	saName := pod.Spec.ServiceAccountName
	if saName == "" {
		saName = "default"
	}
	pod.Spec.ImagePullSecrets = dropGeneratedPullSecrets(pod.Spec.ImagePullSecrets, saName)

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

	// Preserve intentionally set fields
	// Only clear these if they're the default values
	if svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeCluster {
		svc.Spec.ExternalTrafficPolicy = ""
	}

	if svc.Spec.SessionAffinity == v1.ServiceAffinityNone {
		svc.Spec.SessionAffinity = ""
		svc.Spec.SessionAffinityConfig = nil
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

	// The old token is bound to the old service account UID and is rejected after a restore;
	// with empty data the token controller fills in a valid token, CA and namespace
	if secret.Type == v1.SecretTypeServiceAccountToken {
		secret.Data = nil
	}

	// Keep the data and type fields intact for all other Secrets
}

// CleanPVC removes server-side fields from a PersistentVolumeClaim
func CleanPVC(pvc *v1.PersistentVolumeClaim) {
	// A volume picked by the binder (not pre-bound by the user) won't exist after a restore,
	// and the claim would stay Lost; let the restore provision or bind a new one
	if pvc.Annotations["pv.kubernetes.io/bound-by-controller"] == "yes" {
		pvc.Spec.VolumeName = ""
	}

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
	sa.Secrets = nil // Server manages the secrets references
	sa.ImagePullSecrets = dropGeneratedPullSecrets(sa.ImagePullSecrets, sa.Name)

	// Keep user-set fields like imagePullSecrets and automountServiceAccountToken
}

// dropGeneratedPullSecrets removes the <sa>-dockercfg-<random> pull secret OpenShift generates and injects into
// ServiceAccounts and Pods. The secret is skipped from backups and regenerated with a new name on restore.
// shortcut: matches by name, so a user secret named <sa>-dockercfg-* is dropped too; match the owning annotation if that bites.
func dropGeneratedPullSecrets(refs []v1.LocalObjectReference, saName string) []v1.LocalObjectReference {
	return slices.DeleteFunc(refs, func(r v1.LocalObjectReference) bool {
		return strings.HasPrefix(r.Name, saName+"-dockercfg-")
	})
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

// CleanRole removes server-side fields from a Role
func CleanRole(role *rbacv1.Role) {
	// Clean metadata
	CleanMetadata(&role.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	role.APIVersion = "rbac.authorization.k8s.io/v1"
	role.Kind = "Role"
}

// CleanRoleBinding removes server-side fields from a RoleBinding
func CleanRoleBinding(roleBinding *rbacv1.RoleBinding) {
	// Clean metadata
	CleanMetadata(&roleBinding.ObjectMeta)

	// Set API version and kind for valid Kubernetes manifests
	roleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	roleBinding.Kind = "RoleBinding"
}

// CleanRoute removes server-side fields from an OpenShift Route
func CleanRoute(route *routev1.Route) {
	// Remove status (contains router-assigned ingress info)
	route.Status = routev1.RouteStatus{}

	// Clean metadata
	CleanMetadata(&route.ObjectMeta)

	// A server-generated host embeds the source cluster's domain; let the target cluster regenerate it
	if route.Annotations["openshift.io/host.generated"] == "true" {
		route.Spec.Host = ""
		delete(route.Annotations, "openshift.io/host.generated")
	}

	// Set API version and kind for valid OpenShift manifests
	route.APIVersion = "route.openshift.io/v1"
	route.Kind = "Route"
}

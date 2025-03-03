package utils

import (
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsSystemAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"kubernetes.io prefix", "kubernetes.io/service-account-name", true},
		{"k8s.io prefix", "k8s.io/generated-by", true},
		{"app.kubernetes.io prefix", "app.kubernetes.io/name", true},
		{"kubectl.kubernetes.io prefix", "kubectl.kubernetes.io/last-applied-configuration", true},
		{"meta.helm.sh prefix", "meta.helm.sh/release-name", true},
		{"custom annotation", "custom.annotation/value", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSystemAnnotation(tt.key)
			if got != tt.expected {
				t.Errorf("IsSystemAnnotation(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestCleanMetadata(t *testing.T) {
	// Setup a metadata object with fields that should be cleaned
	now := metav1.Now()
	meta := metav1.ObjectMeta{
		Name:                       "test-obj",
		Namespace:                  "test-namespace",
		Labels:                     map[string]string{"app": "test"},
		Annotations:                map[string]string{"custom/ann": "keep", "kubernetes.io/ann": "remove"},
		CreationTimestamp:          metav1.Time{Time: time.Now()},
		DeletionTimestamp:          &now,
		DeletionGracePeriodSeconds: func() *int64 { i := int64(30); return &i }(),
		Generation:                 123,
		ResourceVersion:            "999",
		SelfLink:                   "/api/v1/namespaces/test/pods/test-pod",
		UID:                        "abcd-1234-efgh-5678",
		Finalizers:                 []string{"foregroundDeletion"},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "owner-deploy",
				UID:        "owner-uid",
			},
		},
	}

	// Make a copy to clean
	metaCopy := meta.DeepCopy()
	CleanMetadata(metaCopy)

	// Check that cleaning happened correctly
	if metaCopy.Name != "test-obj" {
		t.Errorf("Name was changed, got %q, want %q", metaCopy.Name, "test-obj")
	}
	if metaCopy.Namespace != "test-namespace" {
		t.Errorf("Namespace was changed, got %q, want %q", metaCopy.Namespace, "test-namespace")
	}
	if !reflect.DeepEqual(metaCopy.Labels, map[string]string{"app": "test"}) {
		t.Errorf("Labels were changed, got %v, want %v", metaCopy.Labels, map[string]string{"app": "test"})
	}
	if !reflect.DeepEqual(metaCopy.Annotations, map[string]string{"custom/ann": "keep"}) {
		t.Errorf("Annotations were not cleaned correctly, got %v, want %v", 
			metaCopy.Annotations, map[string]string{"custom/ann": "keep"})
	}

	// Check that server-side fields were cleared
	if !metaCopy.CreationTimestamp.IsZero() {
		t.Errorf("CreationTimestamp was not zeroed, got %v", metaCopy.CreationTimestamp)
	}
	if metaCopy.DeletionTimestamp != nil {
		t.Errorf("DeletionTimestamp was not cleared, got %v", metaCopy.DeletionTimestamp)
	}
	if metaCopy.DeletionGracePeriodSeconds != nil {
		t.Errorf("DeletionGracePeriodSeconds was not cleared, got %v", *metaCopy.DeletionGracePeriodSeconds)
	}
	if metaCopy.Generation != 0 {
		t.Errorf("Generation was not zeroed, got %v", metaCopy.Generation)
	}
	if metaCopy.ResourceVersion != "" {
		t.Errorf("ResourceVersion was not cleared, got %q", metaCopy.ResourceVersion)
	}
	if metaCopy.SelfLink != "" {
		t.Errorf("SelfLink was not cleared, got %q", metaCopy.SelfLink)
	}
	if string(metaCopy.UID) != "" {
		t.Errorf("UID was not cleared, got %q", metaCopy.UID)
	}
	if len(metaCopy.Finalizers) != 0 {
		t.Errorf("Finalizers were not cleared, got %v", metaCopy.Finalizers)
	}
	if len(metaCopy.OwnerReferences) != 0 {
		t.Errorf("OwnerReferences were not cleared, got %v", metaCopy.OwnerReferences)
	}
}

func TestCleanPod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels:    map[string]string{"app": "test"},
			// Add other metadata fields that should be cleaned
			ResourceVersion: "123",
			UID:             "pod-uid",
		},
		Spec: corev1.PodSpec{
			NodeName:                "node1",
			ServiceAccountName:      "default",
			DeprecatedServiceAccount: "old-sa",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			// Add other status fields
		},
	}

	// Clean the pod
	CleanPod(pod)

	// Verify core fields are preserved
	if pod.Name != "test-pod" {
		t.Errorf("Pod name was changed: got %q, want %q", pod.Name, "test-pod")
	}
	if pod.Namespace != "test-namespace" {
		t.Errorf("Pod namespace was changed: got %q, want %q", pod.Namespace, "test-namespace")
	}
	if !reflect.DeepEqual(pod.Labels, map[string]string{"app": "test"}) {
		t.Errorf("Pod labels were changed: got %v, want %v", pod.Labels, map[string]string{"app": "test"})
	}

	// Verify runtime fields are cleaned
	if pod.ResourceVersion != "" {
		t.Errorf("ResourceVersion not cleared: got %q", pod.ResourceVersion)
	}
	if string(pod.UID) != "" {
		t.Errorf("UID not cleared: got %q", pod.UID)
	}
	if pod.Spec.NodeName != "" {
		t.Errorf("NodeName not cleared: got %q", pod.Spec.NodeName)
	}
	if pod.Spec.DeprecatedServiceAccount != "" {
		t.Errorf("DeprecatedServiceAccount not cleared: got %q", pod.Spec.DeprecatedServiceAccount)
	}
	if pod.Spec.ServiceAccountName != "" {
		t.Errorf("ServiceAccountName 'default' not cleared: got %q", pod.Spec.ServiceAccountName)
	}
	if pod.Status.Phase != "" {
		t.Errorf("Status.Phase not cleared: got %v", pod.Status.Phase)
	}

	// Verify API version and kind are set properly
	if pod.APIVersion != "v1" {
		t.Errorf("APIVersion not set correctly: got %q, want %q", pod.APIVersion, "v1")
	}
	if pod.Kind != "Pod" {
		t.Errorf("Kind not set correctly: got %q, want %q", pod.Kind, "Pod")
	}
}

func TestCleanService(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "10.0.0.1",
			ClusterIPs:               []string{"10.0.0.1"},
			ExternalIPs:              []string{"1.2.3.4"},
			LoadBalancerIP:           "5.6.7.8",
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyTypeCluster,
			SessionAffinity:          corev1.ServiceAffinityNone,
			HealthCheckNodePort:      30000,
			PublishNotReadyAddresses: false,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "9.10.11.12"},
				},
			},
		},
	}

	// Clean the service
	CleanService(svc)

	// Verify core fields are preserved
	if svc.Name != "test-service" {
		t.Errorf("Service name was changed: got %q, want %q", svc.Name, "test-service")
	}
	if svc.Namespace != "test-namespace" {
		t.Errorf("Service namespace was changed: got %q, want %q", svc.Namespace, "test-namespace")
	}

	// Verify cluster-specific fields are cleaned
	if svc.Spec.ClusterIP != "" {
		t.Errorf("ClusterIP not cleared: got %q", svc.Spec.ClusterIP)
	}
	if svc.Spec.ClusterIPs != nil {
		t.Errorf("ClusterIPs not cleared: got %v", svc.Spec.ClusterIPs)
	}
	if svc.Spec.ExternalIPs != nil {
		t.Errorf("ExternalIPs not cleared: got %v", svc.Spec.ExternalIPs)
	}
	if svc.Spec.LoadBalancerIP != "" {
		t.Errorf("LoadBalancerIP not cleared: got %q", svc.Spec.LoadBalancerIP)
	}
	if svc.Spec.ExternalTrafficPolicy != "" {
		t.Errorf("ExternalTrafficPolicy not cleared: got %v", svc.Spec.ExternalTrafficPolicy)
	}
	if svc.Spec.SessionAffinity != "" {
		t.Errorf("SessionAffinity not cleared: got %v", svc.Spec.SessionAffinity)
	}
	if svc.Spec.HealthCheckNodePort != 0 {
		t.Errorf("HealthCheckNodePort not cleared: got %d", svc.Spec.HealthCheckNodePort)
	}

	// Verify status is cleaned
	if !reflect.DeepEqual(svc.Status, corev1.ServiceStatus{}) {
		t.Errorf("Status not cleared: got %v", svc.Status)
	}

	// Verify API version and kind are set properly
	if svc.APIVersion != "v1" {
		t.Errorf("APIVersion not set correctly: got %q, want %q", svc.APIVersion, "v1")
	}
	if svc.Kind != "Service" {
		t.Errorf("Kind not set correctly: got %q, want %q", svc.Kind, "Service")
	}

	// Test a headless service (ClusterIP = None)
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "headless-service",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
		},
	}

	CleanService(headlessSvc)
	if headlessSvc.Spec.ClusterIP != "None" {
		t.Errorf("Headless service ClusterIP should remain 'None', got %q", headlessSvc.Spec.ClusterIP)
	}
}

func TestCleanDeployment(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 3,
		},
	}

	// Clean the deployment
	CleanDeployment(deploy)

	// Verify core fields are preserved
	if deploy.Name != "test-deploy" {
		t.Errorf("Deployment name was changed: got %q, want %q", deploy.Name, "test-deploy")
	}
	if deploy.Namespace != "test-namespace" {
		t.Errorf("Deployment namespace was changed: got %q, want %q", deploy.Namespace, "test-namespace")
	}

	// Verify template labels are preserved
	if !reflect.DeepEqual(deploy.Spec.Template.Labels, map[string]string{"app": "test"}) {
		t.Errorf("Template labels changed: got %v, want %v", deploy.Spec.Template.Labels, map[string]string{"app": "test"})
	}

	// Verify selector is preserved
	if !reflect.DeepEqual(deploy.Spec.Selector.MatchLabels, map[string]string{"app": "test"}) {
		t.Errorf("Selector changed: got %v, want %v", deploy.Spec.Selector.MatchLabels, map[string]string{"app": "test"})
	}

	// Verify status is cleaned
	if !reflect.DeepEqual(deploy.Status, appsv1.DeploymentStatus{}) {
		t.Errorf("Status not cleared: got %v", deploy.Status)
	}

	// Verify API version and kind are set properly
	if deploy.APIVersion != "apps/v1" {
		t.Errorf("APIVersion not set correctly: got %q, want %q", deploy.APIVersion, "apps/v1")
	}
	if deploy.Kind != "Deployment" {
		t.Errorf("Kind not set correctly: got %q, want %q", deploy.Kind, "Deployment")
	}

	// Test a deployment with no selector
	deployNoSelector := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deploy-no-selector",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
			},
			// No selector
		},
	}

	CleanDeployment(deployNoSelector)
	
	// Selector should be created based on template labels
	if deployNoSelector.Spec.Selector == nil {
		t.Errorf("Selector was not created for deployment without selector")
	} else if !reflect.DeepEqual(deployNoSelector.Spec.Selector.MatchLabels, map[string]string{"app": "test"}) {
		t.Errorf("Selector not created correctly: got %v, want %v", 
			deployNoSelector.Spec.Selector.MatchLabels, map[string]string{"app": "test"})
	}
}

func TestCleanConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-configmap",
			Namespace:       "test-namespace",
			ResourceVersion: "123",
			UID:             "cm-uid",
			Annotations: map[string]string{
				"custom/annotation":           "keep-this",
				"kubernetes.io/annotation":    "remove-this",
			},
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		BinaryData: map[string][]byte{
			"binary1": []byte{1, 2, 3, 4},
		},
	}

	// Clean the ConfigMap
	CleanConfigMap(cm)

	// Verify core fields are preserved
	if cm.Name != "test-configmap" {
		t.Errorf("ConfigMap name was changed: got %q, want %q", cm.Name, "test-configmap")
	}
	if cm.Namespace != "test-namespace" {
		t.Errorf("ConfigMap namespace was changed: got %q, want %q", cm.Namespace, "test-namespace")
	}

	// Verify data fields are preserved
	if !reflect.DeepEqual(cm.Data, map[string]string{"key1": "value1", "key2": "value2"}) {
		t.Errorf("ConfigMap data was changed: got %v", cm.Data)
	}
	if !reflect.DeepEqual(cm.BinaryData, map[string][]byte{"binary1": {1, 2, 3, 4}}) {
		t.Errorf("ConfigMap binary data was changed: got %v", cm.BinaryData)
	}

	// Verify runtime fields are cleaned
	if cm.ResourceVersion != "" {
		t.Errorf("ResourceVersion not cleared: got %q", cm.ResourceVersion)
	}
	if string(cm.UID) != "" {
		t.Errorf("UID not cleared: got %q", cm.UID)
	}

	// Verify annotations are cleaned properly
	expectedAnnotations := map[string]string{"custom/annotation": "keep-this"}
	if !reflect.DeepEqual(cm.Annotations, expectedAnnotations) {
		t.Errorf("Annotations not properly cleaned: got %v, want %v", cm.Annotations, expectedAnnotations)
	}

	// Verify API version and kind are set properly
	if cm.APIVersion != "v1" {
		t.Errorf("APIVersion not set correctly: got %q, want %q", cm.APIVersion, "v1")
	}
	if cm.Kind != "ConfigMap" {
		t.Errorf("Kind not set correctly: got %q, want %q", cm.Kind, "ConfigMap")
	}
}
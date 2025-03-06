package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractName(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "Pod with name",
			input: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod",
				},
			},
			expected: "test-pod",
		},
		{
			name: "Unstructured object with name",
			input: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-unstructured",
				},
			},
			expected: "test-unstructured",
		},
		{
			name:     "Nil object",
			input:    nil,
			expected: "",
		},
		{
			name:     "Object without metadata",
			input:    "not-a-k8s-object",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractName(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractItems(t *testing.T) {
	// Test with PodList
	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod2",
				},
			},
		},
	}

	items, count := ExtractItems(podList)
	if count != 2 {
		t.Errorf("ExtractItems(PodList) count = %v, want %v", count, 2)
	}
	if len(items) != 2 {
		t.Errorf("ExtractItems(PodList) items length = %v, want %v", len(items), 2)
	}

	// Test with DeploymentList
	deployList := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deploy1",
				},
			},
		},
	}

	items, count = ExtractItems(deployList)
	if count != 1 {
		t.Errorf("ExtractItems(DeploymentList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(DeploymentList) items length = %v, want %v", len(items), 1)
	}

	// Test with unstructured object
	unstructured := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "item1"},
			map[string]interface{}{"name": "item2"},
		},
	}

	items, count = ExtractItems(unstructured)
	if count != 2 {
		t.Errorf("ExtractItems(unstructured) count = %v, want %v", count, 2)
	}
	if len(items) != 2 {
		t.Errorf("ExtractItems(unstructured) items length = %v, want %v", len(items), 2)
	}

	// Test with empty list
	emptyList := &corev1.PodList{}
	items, count = ExtractItems(emptyList)
	if count != 0 {
		t.Errorf("ExtractItems(emptyList) count = %v, want %v", count, 0)
	}
	if len(items) != 0 {
		t.Errorf("ExtractItems(emptyList) items length = %v, want %v", len(items), 0)
	}

	// Test with ServiceList
	serviceList := &corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "service1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "service2",
				},
			},
		},
	}
	items, count = ExtractItems(serviceList)
	if count != 2 {
		t.Errorf("ExtractItems(ServiceList) count = %v, want %v", count, 2)
	}
	if len(items) != 2 {
		t.Errorf("ExtractItems(ServiceList) items length = %v, want %v", len(items), 2)
	}

	// Test with ConfigMapList
	configMapList := &corev1.ConfigMapList{
		Items: []corev1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "config1",
				},
			},
		},
	}
	items, count = ExtractItems(configMapList)
	if count != 1 {
		t.Errorf("ExtractItems(ConfigMapList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(ConfigMapList) items length = %v, want %v", len(items), 1)
	}

	// Test with SecretList
	secretList := &corev1.SecretList{
		Items: []corev1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "secret1",
				},
			},
		},
	}
	items, count = ExtractItems(secretList)
	if count != 1 {
		t.Errorf("ExtractItems(SecretList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(SecretList) items length = %v, want %v", len(items), 1)
	}

	// Test with PersistentVolumeClaimList
	pvcList := &corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc1",
				},
			},
		},
	}
	items, count = ExtractItems(pvcList)
	if count != 1 {
		t.Errorf("ExtractItems(PersistentVolumeClaimList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(PersistentVolumeClaimList) items length = %v, want %v", len(items), 1)
	}

	// Test with ServiceAccountList
	saList := &corev1.ServiceAccountList{
		Items: []corev1.ServiceAccount{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sa1",
				},
			},
		},
	}
	items, count = ExtractItems(saList)
	if count != 1 {
		t.Errorf("ExtractItems(ServiceAccountList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(ServiceAccountList) items length = %v, want %v", len(items), 1)
	}

	// Test with StatefulSetList
	statefulSetList := &appsv1.StatefulSetList{
		Items: []appsv1.StatefulSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "statefulset1",
				},
			},
		},
	}
	items, count = ExtractItems(statefulSetList)
	if count != 1 {
		t.Errorf("ExtractItems(StatefulSetList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(StatefulSetList) items length = %v, want %v", len(items), 1)
	}

	// Test with DaemonSetList
	daemonSetList := &appsv1.DaemonSetList{
		Items: []appsv1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "daemonset1",
				},
			},
		},
	}
	items, count = ExtractItems(daemonSetList)
	if count != 1 {
		t.Errorf("ExtractItems(DaemonSetList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(DaemonSetList) items length = %v, want %v", len(items), 1)
	}

	// Test with IngressList
	ingressList := &networkingv1.IngressList{
		Items: []networkingv1.Ingress{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ingress1",
				},
			},
		},
	}
	items, count = ExtractItems(ingressList)
	if count != 1 {
		t.Errorf("ExtractItems(IngressList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(IngressList) items length = %v, want %v", len(items), 1)
	}

	// Test with RoleList
	roleList := &rbacv1.RoleList{
		Items: []rbacv1.Role{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "role1",
				},
			},
		},
	}
	items, count = ExtractItems(roleList)
	if count != 1 {
		t.Errorf("ExtractItems(RoleList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(RoleList) items length = %v, want %v", len(items), 1)
	}

	// Test with RoleBindingList
	roleBindingList := &rbacv1.RoleBindingList{
		Items: []rbacv1.RoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rolebinding1",
				},
			},
		},
	}
	items, count = ExtractItems(roleBindingList)
	if count != 1 {
		t.Errorf("ExtractItems(RoleBindingList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(RoleBindingList) items length = %v, want %v", len(items), 1)
	}

	// Test with CronJobList
	cronJobList := &batchv1.CronJobList{
		Items: []batchv1.CronJob{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cronjob1",
				},
			},
		},
	}
	items, count = ExtractItems(cronJobList)
	if count != 1 {
		t.Errorf("ExtractItems(CronJobList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(CronJobList) items length = %v, want %v", len(items), 1)
	}

	// Test with JobList
	jobList := &batchv1.JobList{
		Items: []batchv1.Job{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "job1",
				},
			},
		},
	}
	items, count = ExtractItems(jobList)
	if count != 1 {
		t.Errorf("ExtractItems(JobList) count = %v, want %v", count, 1)
	}
	if len(items) != 1 {
		t.Errorf("ExtractItems(JobList) items length = %v, want %v", len(items), 1)
	}
}

package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
}
package utils

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSkipReason(t *testing.T) {
	isController := true
	owned := func(kind, name string) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name:            "child",
			OwnerReferences: []metav1.OwnerReference{{Kind: kind, Name: name, Controller: &isController}},
		}
	}
	notController := metav1.ObjectMeta{
		Name:            "child",
		OwnerReferences: []metav1.OwnerReference{{Kind: "ConfigMap", Name: "cm"}},
	}

	tests := []struct {
		name string
		obj  interface{}
		skip bool
	}{
		{"pod owned by ReplicaSet", &corev1.Pod{ObjectMeta: owned("ReplicaSet", "web-abc")}, true},
		{"static mirror pod owned by Node", &corev1.Pod{ObjectMeta: owned("Node", "cplane-01")}, true},
		{"bare pod", &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "bare"}}, false},
		{"pod with non-controller owner", &corev1.Pod{ObjectMeta: notController}, false},
		{"job owned by CronJob", &batchv1.Job{ObjectMeta: owned("CronJob", "tick")}, true},
		{"standalone job", &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "once"}}, false},
		{"kube-root-ca.crt", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "kube-root-ca.crt"}}, true},
		{"user configmap", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app-config"}}, false},
		{"other kinds with a controller are kept", &corev1.Service{ObjectMeta: owned("Operator", "x")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SkipReason(tt.obj) != ""; got != tt.skip {
				t.Errorf("SkipReason() skip = %v, want %v (reason %q)", got, tt.skip, SkipReason(tt.obj))
			}
		})
	}
}

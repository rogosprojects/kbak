package utils

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	secret := func(name, annotation, value string) *corev1.Secret {
		return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: map[string]string{annotation: value}}}
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

		// OpenShift
		{"openshift-service-ca.crt", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "openshift-service-ca.crt"}}, true},
		{"4.16+ generated pull secret", secret("builder-dockercfg-x7k2p", "openshift.io/internal-registry-auth-token.service-account", "builder"), true},
		{"legacy generated dockercfg", secret("default-dockercfg-abcde", "openshift.io/token-secret.name", "default-token-fghij"), true},
		{"legacy generated SA token", secret("default-token-fghij", "kubernetes.io/created-by", "openshift.io/create-dockercfg-secrets"), true},
		{"service-ca serving cert", secret("web-tls", "service.beta.openshift.io/originating-service-name", "web"), true},
		{"user token secret", secret("app-sa-token", "kubernetes.io/service-account.name", "app-sa"), false},
		{"user secret", secret("app-secret", "team/owner", "platform"), false},
		{"image pullers binding", &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "system:image-pullers"}}, true},
		{"deployers binding", &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "system:deployers"}}, true},
		{"project admin binding", &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "admin"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SkipReason(tt.obj) != ""; got != tt.skip {
				t.Errorf("SkipReason() skip = %v, want %v (reason %q)", got, tt.skip, SkipReason(tt.obj))
			}
		})
	}
}

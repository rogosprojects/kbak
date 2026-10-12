package utils

import (
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SkipReason returns why obj should be left out of a backup, or "" to back it up.
// These objects are recreated by the cluster; restoring them adds orphaned duplicates.
func SkipReason(obj interface{}) string {
	switch o := obj.(type) {
	case *v1.Pod, *batchv1.Job:
		// Pods of Deployments/StatefulSets/DaemonSets/Jobs, static mirror pods, and Jobs of CronJobs
		if ref := metav1.GetControllerOf(o.(metav1.Object)); ref != nil {
			return "managed by " + ref.Kind + "/" + ref.Name
		}
	case *v1.ConfigMap:
		if o.Name == "kube-root-ca.crt" {
			return "published by Kubernetes in every namespace"
		}
	}
	return ""
}

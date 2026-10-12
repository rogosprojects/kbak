package utils

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ExtractItems returns pointers to the items of any typed Kubernetes list (e.g. *v1.PodList),
// and their count. Anything that is not a list yields no items.
func ExtractItems(list interface{}) ([]interface{}, int) {
	obj, ok := list.(runtime.Object)
	if !ok {
		return nil, 0
	}
	objects, err := meta.ExtractList(obj)
	if err != nil {
		return nil, 0
	}
	items := make([]interface{}, len(objects))
	for i, o := range objects {
		items[i] = o
	}
	return items, len(items)
}

// ExtractName attempts to get the name of a Kubernetes resource
func ExtractName(obj interface{}) string {
	if metaObj, ok := obj.(metav1.Object); ok {
		return metaObj.GetName()
	}
	return ""
}

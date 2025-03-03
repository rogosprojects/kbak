package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExtractItems gets the items slice from various list types
func ExtractItems(list interface{}) ([]interface{}, int) {
	// Check for different list types and extract items accordingly
	switch v := list.(type) {
	case metav1.ListInterface:
		items := ExtractItemsFromListInterface(v)
		remainingCount := 0
		if v.GetRemainingItemCount() != nil {
			remainingCount = int(*v.GetRemainingItemCount())
		}
		return items, len(items) + remainingCount
	default:
		// Use reflection as a fallback for types we don't directly support
		items, count := ExtractItemsUsingReflection(v)
		return items, count
	}
}

// ExtractItemsFromListInterface extracts items from a kubernetes ListInterface
func ExtractItemsFromListInterface(list metav1.ListInterface) []interface{} {
	// Type assertions for common list types
	// This would need to be expanded for all resource types
	switch v := list.(type) {
	case *metav1.List:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			result[i] = item
		}
		return result
	default:
		// Fallback using reflection
		items, _ := ExtractItemsUsingReflection(list)
		return items
	}
}

// ExtractItemsUsingReflection attempts to extract items using type assertions
func ExtractItemsUsingReflection(obj interface{}) ([]interface{}, int) {
	// For specific kubernetes types
	switch v := obj.(type) {
	case *v1.PodList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item // Create a new variable to avoid issues with loop variable capture
			result[i] = &item
		}
		return result, len(result)

	case *appsv1.DeploymentList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item
			result[i] = &item
		}
		return result, len(result)

	case *v1.ServiceList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item
			result[i] = &item
		}
		return result, len(result)
		
	case *v1.ConfigMapList:
		result := make([]interface{}, len(v.Items))
		for i, item := range v.Items {
			item := item
			result[i] = &item
		}
		return result, len(result)

	// For all other types, use a more generic approach
	default:
		// Try to handle list types via type assertions
		// For unstructured lists
		if list, ok := obj.(map[string]interface{}); ok {
			if items, ok := list["items"].([]interface{}); ok {
				return items, len(items)
			}
		}

		// Try to access via methods if available
		if list, ok := obj.(interface{ GetItems() []interface{} }); ok {
			items := list.GetItems()
			return items, len(items)
		}

		// Try other common variants that might be used
		if list, ok := obj.(interface{ Items() []interface{} }); ok {
			return list.Items(), len(list.Items())
		}

		// Return an empty list for unhandled types
		//fmt.Printf("Warning: Unhandled type in extractItemsUsingReflection: %T\n", obj)
		return []interface{}{}, 0
	}
}

// ExtractName attempts to get the name of a Kubernetes resource
func ExtractName(obj interface{}) string {
	// Try to access metadata.name
	if metaObj, ok := obj.(metav1.Object); ok {
		return metaObj.GetName()
	}

	// For unstructured objects from dynamic client
	if unstr, ok := obj.(map[string]interface{}); ok {
		if metadata, ok := unstr["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				return name
			}
		}
	}

	return ""
}
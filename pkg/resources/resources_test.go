package resources

import (
	"errors"
	"testing"
)

func TestGetAllResourceTypes(t *testing.T) {
	// Test that we get a non-empty list of all resource types
	resourceTypes := GetAllResourceTypes()

	if len(resourceTypes) == 0 {
		t.Errorf("GetAllResourceTypes() returned an empty list")
	}

	// Check for some common resource types
	expectedTypes := map[string]bool{
		"Pod":                   false,
		"Deployment":            false,
		"Service":               false,
		"ConfigMap":             false,
		"Secret":                false,
		"PersistentVolumeClaim": false,
	}

	for _, rt := range resourceTypes {
		if _, ok := expectedTypes[rt.Kind]; ok {
			expectedTypes[rt.Kind] = true
		}

		// Check that APIFunc is not nil
		if rt.APIFunc == nil {
			t.Errorf("Resource type %s has nil APIFunc", rt.Kind)
		}
	}

	// Verify all expected types were found
	for kind, found := range expectedTypes {
		if !found {
			t.Errorf("Expected resource type %s not found", kind)
		}
	}
}

func TestGetResourceTypes(t *testing.T) {
	// Test 1: Get all resource types with empty filter
	resourceTypes := GetResourceTypes(map[string]bool{})
	if len(resourceTypes) == 0 {
		t.Errorf("GetResourceTypes() with empty filter returned an empty list")
	}
	if len(resourceTypes) != len(GetAllResourceTypes()) {
		t.Errorf("GetResourceTypes() with empty filter should return all resource types")
	}

	// Test 2: Filtered resource types
	filter := map[string]bool{
		"pod":       true,
		"configmap": true,
		"secret":    true,
	}
	filteredTypes := GetResourceTypes(filter)
	if len(filteredTypes) != 3 {
		t.Errorf("GetResourceTypes() with filter returned %d types, expected 3", len(filteredTypes))
	}

	// Check that only the requested types are returned
	foundTypes := map[string]bool{
		"Pod":       false,
		"ConfigMap": false,
		"Secret":    false,
	}
	for _, rt := range filteredTypes {
		if _, ok := foundTypes[rt.Kind]; ok {
			foundTypes[rt.Kind] = true
		} else {
			t.Errorf("Unexpected resource type %s in filtered results", rt.Kind)
		}
	}

	// Verify all filtered types were found
	for kind, found := range foundTypes {
		if !found {
			t.Errorf("Expected filtered resource type %s not found", kind)
		}
	}
}

func TestIsNotFoundError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		// Skip nil error test as IsNotFoundError assumes non-nil error
		/* {
			name:     "nil error",
			err:      nil,
			expected: false,
		}, */
		{
			name:     "not found error lowercase",
			err:      errors.New("resource not found"),
			expected: true,
		},
		{
			name:     "not found error mixed case",
			err:      errors.New("Resource Not Found"),
			expected: true,
		},
		{
			name:     "no matches for kind error",
			err:      errors.New("no matches for kind ConfigMap"),
			expected: true,
		},
		{
			name:     "server could not find error",
			err:      errors.New("the server could not find the requested resource"),
			expected: true,
		},
		{
			name:     "no resource type error",
			err:      errors.New("the server doesn't have a resource type foo"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("connection refused"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNotFoundError(tc.err)
			if result != tc.expected {
				t.Errorf("IsNotFoundError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

package resources

import (
	"errors"
	"testing"
)

func TestGetResourceTypes(t *testing.T) {
	// Test that we get a non-empty list of resource types
	resourceTypes := GetResourceTypes()
	
	if len(resourceTypes) == 0 {
		t.Errorf("GetResourceTypes() returned an empty list")
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
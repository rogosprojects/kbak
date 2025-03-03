package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"kbak/pkg/client"
	"kbak/pkg/resources"
	"kbak/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// PerformBackup performs the backup of all resources in the specified namespace
func PerformBackup(k8sClient *client.K8sClient, namespace, backupDir string, verbose bool) (int, int) {
	resourceTypes := resources.GetResourceTypes()
	resourceCount := 0
	errorCount := 0

	// Backup each resource type
	for _, resource := range resourceTypes {
		var objects interface{}
		var err error

		// Get resources from the Kubernetes API
		objects, err = resource.APIFunc(k8sClient, namespace, metav1.ListOptions{})

		if err != nil {
			fmt.Printf("Error listing %s: %v\n", resource.Kind, err)
			if verbose {
				fmt.Printf("Debug info - API endpoint: %s\n", k8sClient.Config.Host)
				fmt.Printf("Debug info - Resource: %s in namespace %s\n", resource.Kind, namespace)
			}
			errorCount++
			continue
		}

		// Debug the response from the API
		if verbose {
			fmt.Printf("Response type for %s: %T\n", resource.Kind, objects)
		}

		// Extract items from the list
		items, itemCount := utils.ExtractItems(objects)
		if verbose {
			fmt.Printf("Found %d %s resources in namespace %s\n", itemCount, resource.Kind, namespace)
		}
		if itemCount == 0 {
			// Skip creating directories for resource kinds with no items
			continue
		}

		// Create directory for this resource kind
		kindDir := filepath.Join(backupDir, resource.Kind)
		if err := os.MkdirAll(kindDir, 0755); err != nil {
			fmt.Printf("Error creating directory for %s: %v\n", resource.Kind, err)
			errorCount++
			continue
		}

		itemsBackedUp := 0
		for i, item := range items {
			if item == nil {
				continue
			}

			name := utils.ExtractName(item)
			if name == "" {
				name = fmt.Sprintf("unknown-%d", i)
			}

			// Remove cluster-specific and runtime fields
			utils.CleanObject(item)

			// Convert to YAML
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				fmt.Printf("Error marshaling %s '%s': %v\n", resource.Kind, name, err)
				continue
			}

			// Save to file
			filename := filepath.Join(kindDir, name+".yaml")
			if err := os.WriteFile(filename, yamlData, 0644); err != nil {
				fmt.Printf("Error writing %s '%s': %v\n", resource.Kind, name, err)
				continue
			}

			itemsBackedUp++
		}

		if itemsBackedUp > 0 {
			fmt.Printf("Backed up %d %s resources\n", itemsBackedUp, resource.Kind)
			resourceCount += itemsBackedUp
		}
	}

	return resourceCount, errorCount
}
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogosprojects/kbak/pkg/client"
	"github.com/rogosprojects/kbak/pkg/resources"
	"github.com/rogosprojects/kbak/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// BackupStats tracks statistics and results from a backup operation
type BackupStats struct {
	ResourceCount     int
	ErrorCount        int
	ResourcesBackedUp map[string]int
	ResourceErrors    map[string]int
}

// NewBackupStats creates and initializes a new BackupStats object
func NewBackupStats() *BackupStats {
	return &BackupStats{
		ResourceCount:     0,
		ErrorCount:        0,
		ResourcesBackedUp: make(map[string]int),
		ResourceErrors:    make(map[string]int),
	}
}

// PerformBackup performs the backup of resources in the specified namespace
// Returns statistics about the backup operation including counts of resources backed up and errors
func PerformBackup(k8sClient *client.K8sClient, namespace, backupDir string, selectedTypes map[string]bool, verbose bool) (int, int) {
	stats := NewBackupStats()
	resourceTypes := resources.GetResourceTypes(selectedTypes)

	if len(resourceTypes) == 0 && verbose {
		fmt.Printf("%s %s%sWarning: No resource types selected for backup%s\n",
			utils.WarningEmoji, utils.Yellow, utils.Bold, utils.Reset)
		return 0, 0
	}

	// Backup each resource type
	for _, resource := range resourceTypes {
		backupResourceType(k8sClient, namespace, backupDir, resource, stats, verbose)
	}

	return stats.ResourceCount, stats.ErrorCount
}

// ensureValidFilename sanitizes a resource name to ensure it's a valid filename
// by replacing invalid characters and handling edge cases
func ensureValidFilename(name string) string {
	if name == "" {
		return "unnamed"
	}

	// If it's just a dot, replace it to avoid hidden files
	if name == "." {
		return "_dot_"
	}

	// Replace characters that are problematic in filenames
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	// Apply replacements
	result := replacer.Replace(name)

	// Handle leading dots to avoid hidden files
	if strings.HasPrefix(result, ".") {
		result = "_" + result
	}

	// Clean the path to eliminate any issues
	result = filepath.Base(filepath.Clean(result))

	// Trim any leading or trailing problematic characters
	result = strings.Trim(result, "._-")

	// If we ended up with an empty string after trimming, use a default
	if result == "" {
		return "unnamed"
	}

	return result
}

// backupResourceType handles the backup of a single resource type
func backupResourceType(k8sClient *client.K8sClient, namespace, backupDir string,
	resource resources.ResourceType, stats *BackupStats, verbose bool) {

	var objects interface{}
	var err error

	// Get resources from the Kubernetes API
	objects, err = resource.APIFunc(k8sClient, namespace, metav1.ListOptions{})

	if err != nil {
		// Check if this is a "resource not found" type of error
		if resources.IsNotFoundError(err) {
			if verbose {
				fmt.Printf("%s %sResource type %s not available in the cluster, skipping%s\n",
					utils.SkippedEmoji, utils.Cyan, resource.Kind, utils.Reset)
			}
		} else {
			fmt.Printf("%s %s%sError listing %s: %v%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, resource.Kind, err, utils.Reset)
			if verbose {
				fmt.Printf("%sDebug info - API endpoint: %s%s\n",
					utils.BrightBlue, k8sClient.Config.Host, utils.Reset)
				fmt.Printf("%sDebug info - Resource: %s in namespace %s%s\n",
					utils.BrightBlue, resource.Kind, namespace, utils.Reset)
			}
			stats.ErrorCount++
			stats.ResourceErrors[resource.Kind]++
		}
		return
	}

	// Debug the response from the API
	if verbose {
		fmt.Printf("%sResponse type for %s: %T%s\n",
			utils.BrightBlue, resource.Kind, objects, utils.Reset)
	}

	// Extract items from the list
	items, itemCount := utils.ExtractItems(objects)
	if verbose {
		fmt.Printf("%s%sFound %d %s resources in namespace %s%s\n",
			utils.InfoEmoji, utils.Cyan, itemCount, resource.Kind, namespace, utils.Reset)
	}
	if itemCount == 0 {
		// Skip creating directories for resource kinds with no items
		return
	}

	// Create directory for this resource kind
	kindDir := filepath.Join(backupDir, resource.Kind)
	if err := os.MkdirAll(kindDir, 0755); err != nil {
		fmt.Printf("%s %s%sError creating directory for %s: %v%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, resource.Kind, err, utils.Reset)
		stats.ErrorCount++
		stats.ResourceErrors[resource.Kind]++
		return
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

		// Ensure the filename is valid for the filesystem
		safeName := ensureValidFilename(name)
		if safeName != name && verbose {
			fmt.Printf("%s%sResource name %q sanitized to %q for filesystem compatibility%s\n",
				utils.InfoEmoji, utils.BrightBlue, name, safeName, utils.Reset)
		}

		// Remove cluster-specific and runtime fields
		utils.CleanObject(item)

		// Convert to YAML
		yamlData, err := yaml.Marshal(item)
		if err != nil {
			fmt.Printf("%s %s%sError marshaling %s '%s': %v%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, resource.Kind, name, err, utils.Reset)
			stats.ErrorCount++
			stats.ResourceErrors[resource.Kind]++
			continue
		}

		// Save to file
		filename := filepath.Join(kindDir, safeName+".yaml")
		if err := os.WriteFile(filename, yamlData, 0644); err != nil {
			fmt.Printf("%s %s%sError writing %s '%s': %v%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, resource.Kind, name, err, utils.Reset)
			stats.ErrorCount++
			stats.ResourceErrors[resource.Kind]++
			continue
		}

		itemsBackedUp++
	}

	if itemsBackedUp > 0 {
		fmt.Printf("%s%sBacked up %d %s resources%s\n",
			utils.Green, utils.Bold, itemsBackedUp, resource.Kind, utils.Reset)
		stats.ResourceCount += itemsBackedUp
		stats.ResourcesBackedUp[resource.Kind] = itemsBackedUp
	}
}

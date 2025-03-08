package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rogosprojects/kbak/pkg/backup"
	"github.com/rogosprojects/kbak/pkg/client"
	"github.com/rogosprojects/kbak/pkg/utils"
	"k8s.io/client-go/util/homedir"
)

// Version is the current version of kbak.
// It will be overridden during build when using ldflags.
var Version = "dev"

// Define resource type flags
type resourceFlags struct {
	all            bool
	pod            bool
	deployment     bool
	service        bool
	configmap      bool
	secret         bool
	pvc            bool
	serviceaccount bool
	statefulset    bool
	daemonset      bool
	ingress        bool
	role           bool
	rolebinding    bool
	cronjob        bool
	job            bool
}

func main() {
	var namespace string
	var kubeconfig string
	var outputDir string
	var verbose bool
	var showVersion bool
	var allNamespaces bool

	// Define resource type flags
	var resFlags resourceFlags

	// Basic flags
	flag.StringVar(&namespace, "namespace", "", "Namespace to backup (required unless --all-namespaces is used)")
	flag.StringVar(&outputDir, "output", "backups", "Output directory for backup files")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&allNamespaces, "all-namespaces", false, "Backup resources from all namespaces")

	// Resource type flags
	flag.BoolVar(&resFlags.all, "all-resources", true, "Backup all resource types (default)")
	flag.BoolVar(&resFlags.pod, "pod", false, "Backup only pods")
	flag.BoolVar(&resFlags.deployment, "deployment", false, "Backup only deployments")
	flag.BoolVar(&resFlags.service, "service", false, "Backup only services")
	flag.BoolVar(&resFlags.configmap, "configmap", false, "Backup only configmaps")
	flag.BoolVar(&resFlags.secret, "secret", false, "Backup only secrets")
	flag.BoolVar(&resFlags.pvc, "pvc", false, "Backup only persistent volume claims")
	flag.BoolVar(&resFlags.serviceaccount, "serviceaccount", false, "Backup only service accounts")
	flag.BoolVar(&resFlags.statefulset, "statefulset", false, "Backup only statefulsets")
	flag.BoolVar(&resFlags.daemonset, "daemonset", false, "Backup only daemonsets")
	flag.BoolVar(&resFlags.ingress, "ingress", false, "Backup only ingresses")
	flag.BoolVar(&resFlags.role, "role", false, "Backup only roles")
	flag.BoolVar(&resFlags.rolebinding, "rolebinding", false, "Backup only rolebindings")
	flag.BoolVar(&resFlags.cronjob, "cronjob", false, "Backup only cronjobs")
	flag.BoolVar(&resFlags.job, "job", false, "Backup only jobs")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("%s %skbak%s version %s %s%s\n",
			utils.K8sEmoji, utils.Bold, utils.Reset, utils.Cyan, Version, utils.Reset)
		os.Exit(0)
	}

	// Validate namespace requirements
	if namespace == "" && !allNamespaces {
		fmt.Printf("%s %s%sError: either --namespace or --all-namespaces flag is required%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, utils.Reset)
		flag.Usage()
		os.Exit(1)
	}

	if allNamespaces && namespace != "" {
		fmt.Printf("%s %s%sWarning: --namespace flag is ignored when --all-namespaces is used%s\n",
			utils.WarningEmoji, utils.Yellow, utils.Bold, utils.Reset)
		namespace = ""
	}

	// Initialize Kubernetes client first to validate connectivity
	k8sClient, err := client.NewClient(kubeconfig, verbose)
	if err != nil {
		fmt.Printf("%s %s%sError initializing Kubernetes client: %v%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, err, utils.Reset)
		os.Exit(1)
	}

	// Handle all-namespaces case (not implemented yet, just a placeholder)
	if allNamespaces {
		fmt.Printf("%s %s%sError: --all-namespaces is not implemented yet%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, utils.Reset)
		os.Exit(1)
		// Here we would get a list of all namespaces and iterate through them
		// This feature is left for future implementation
	}

	// Create output directory with timestamp
	timestamp := time.Now().Format("02Jan2006-15:04")
	backupDir := filepath.Join(outputDir, timestamp, namespace)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("%s %s%sError creating output directory: %v%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, err, utils.Reset)
		os.Exit(1)
	}

	// Prepare resource type filter
	selectedTypes := buildResourceTypeMap(resFlags)

	if len(selectedTypes) > 0 {
		fmt.Printf("%s %s%sStarting backup of selected resource types from namespace '%s' to '%s'%s\n\n",
			utils.StartEmoji, utils.Blue, utils.Bold, namespace, backupDir, utils.Reset)
	} else {
		fmt.Printf("%s %s%sStarting backup of all resource types from namespace '%s' to '%s'%s\n\n",
			utils.StartEmoji, utils.Blue, utils.Bold, namespace, backupDir, utils.Reset)
	}

	// Perform backup
	resourceCount, errorCount := backup.PerformBackup(k8sClient, namespace, backupDir, selectedTypes, verbose)

	if resourceCount > 0 {
		fmt.Printf("\n%s %s%sBackup completed successfully to %s (%d resources total)%s\n",
			utils.SuccessEmoji, utils.Green, utils.Bold, backupDir, resourceCount, utils.Reset)
	} else {
		fmt.Printf("\n%s %s%sNo resources found to backup in namespace '%s'%s\n",
			utils.InfoEmoji, utils.Yellow, utils.Bold, namespace, utils.Reset)
	}

	// Exit with error code if there were errors
	if errorCount > 0 {
		fmt.Printf("%s %s%sCompleted with %d errors%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, errorCount, utils.Reset)
		os.Exit(1)
	}
}

// buildResourceTypeMap creates a map of resource types to include in the backup
// If any specific resource type flags are set, only those types are included
// If no specific flags are set (or --all-resources is true), all resource types are included
func buildResourceTypeMap(flags resourceFlags) map[string]bool {
	selectedTypes := make(map[string]bool)

	// Check if any specific resource type flags are set
	specificTypesSelected := flags.pod || flags.deployment || flags.service ||
		flags.configmap || flags.secret || flags.pvc ||
		flags.serviceaccount || flags.statefulset || flags.daemonset ||
		flags.ingress || flags.role || flags.rolebinding ||
		flags.cronjob || flags.job

	// If no specific types are selected or --all-resources is true (default), return empty map to include all
	if flags.all && !specificTypesSelected {
		return selectedTypes
	}

	// Add selected resource types to the map
	if flags.pod {
		selectedTypes["pod"] = true
	}
	if flags.deployment {
		selectedTypes["deployment"] = true
	}
	if flags.service {
		selectedTypes["service"] = true
	}
	if flags.configmap {
		selectedTypes["configmap"] = true
	}
	if flags.secret {
		selectedTypes["secret"] = true
	}
	if flags.pvc {
		selectedTypes["persistentvolumeclaim"] = true
	}
	if flags.serviceaccount {
		selectedTypes["serviceaccount"] = true
	}
	if flags.statefulset {
		selectedTypes["statefulset"] = true
	}
	if flags.daemonset {
		selectedTypes["daemonset"] = true
	}
	if flags.ingress {
		selectedTypes["ingress"] = true
	}
	if flags.role {
		selectedTypes["role"] = true
	}
	if flags.rolebinding {
		selectedTypes["rolebinding"] = true
	}
	if flags.cronjob {
		selectedTypes["cronjob"] = true
	}
	if flags.job {
		selectedTypes["job"] = true
	}

	return selectedTypes
}

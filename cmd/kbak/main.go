package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rogosprojects/kbak/pkg/backup"
	"github.com/rogosprojects/kbak/pkg/client"
	"k8s.io/client-go/util/homedir"
)

// Version is the current version of kbak.
// It will be overridden during build when using ldflags.
var Version = "dev"

func main() {
	var namespace string
	var kubeconfig string
	var outputDir string
	var verbose bool
	var showVersion bool
	var allNamespaces bool

	flag.StringVar(&namespace, "namespace", "", "Namespace to backup (required unless --all-namespaces is used)")
	flag.StringVar(&outputDir, "output", "backups", "Output directory for backup files")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&allNamespaces, "all-namespaces", false, "Backup resources from all namespaces")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("kbak version %s\n", Version)
		os.Exit(0)
	}

	// Validate namespace requirements
	if namespace == "" && !allNamespaces {
		fmt.Println("Error: either --namespace or --all-namespaces flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if allNamespaces && namespace != "" {
		fmt.Println("Warning: --namespace flag is ignored when --all-namespaces is used")
		namespace = ""
	}

	// Initialize Kubernetes client first to validate connectivity
	k8sClient, err := client.NewClient(kubeconfig, verbose)
	if err != nil {
		fmt.Printf("Error initializing Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Handle all-namespaces case (not implemented yet, just a placeholder)
	if allNamespaces {
		fmt.Println("Error: --all-namespaces is not implemented yet")
		os.Exit(1)
		// Here we would get a list of all namespaces and iterate through them
		// This feature is left for future implementation
	}

	// Create output directory with timestamp
	timestamp := time.Now().Format("02Jan2006-15:04")
	backupDir := filepath.Join(outputDir, timestamp, namespace)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting backup of namespace '%s' to '%s'\n", namespace, backupDir)

	// Perform backup
	resourceCount, errorCount := backup.PerformBackup(k8sClient, namespace, backupDir, verbose)

	if resourceCount > 0 {
		fmt.Printf("Backup completed successfully to %s (%d resources total)\n", backupDir, resourceCount)
	} else {
		fmt.Printf("No resources found to backup in namespace '%s'\n", namespace)
	}

	// Exit with error code if there were errors
	if errorCount > 0 {
		fmt.Printf("Completed with %d errors\n", errorCount)
		os.Exit(1)
	}
}

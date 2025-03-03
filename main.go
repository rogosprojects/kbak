package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/util/homedir"
	"kbak/pkg/backup"
	"kbak/pkg/client"
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

	flag.StringVar(&namespace, "namespace", "", "Namespace to backup (required)")
	flag.StringVar(&outputDir, "output", "backups", "Output directory for backup files")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")

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

	if namespace == "" {
		fmt.Println("Error: --namespace flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create output directory with timestamp
	timestamp := time.Now().Format("02Jan2006-15:04")
	backupDir := filepath.Join(outputDir, timestamp, namespace)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting backup of namespace '%s' to '%s'\n", namespace, backupDir)

	// Initialize Kubernetes client
	k8sClient, err := client.NewClient(kubeconfig, verbose)
	if err != nil {
		fmt.Printf("Error initializing Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Perform backup
	resourceCount, errorCount := backup.PerformBackup(k8sClient, namespace, backupDir, verbose)

	if errorCount > 0 {
		fmt.Printf("Completed with %d errors\n", errorCount)
	}

	if resourceCount > 0 {
		fmt.Printf("Backup completed successfully to %s (%d resources total)\n", backupDir, resourceCount)
	} else {
		fmt.Printf("No resources found to backup in namespace '%s'\n", namespace)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/rogosprojects/kbak/pkg/backup"
	"github.com/rogosprojects/kbak/pkg/client"
	"github.com/rogosprojects/kbak/pkg/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	route          bool
}

func main() {
	var namespace string
	var kubeconfig string
	var outputDir string
	var verbose bool
	var showVersion bool
	var allNamespaces bool
	var recipientsFile string

	// Define resource type flags
	var resFlags resourceFlags

	// Basic flags
	flag.StringVar(&namespace, "namespace", "", "Namespace to backup (uses current namespace from kubeconfig if not specified)")
	flag.StringVar(&outputDir, "output", "backups", "Output directory for backup files")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&allNamespaces, "all-namespaces", false, "Backup resources from all namespaces")
	flag.StringVar(&recipientsFile, "age-recipients-file", "", "Encrypt every file with age to the public keys in this file (one per line, as for age -R); writes <name>.yaml.age")

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
	flag.BoolVar(&resFlags.route, "route", false, "Backup only OpenShift routes")

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: $KUBECONFIG, ~/.kube/config, or in-cluster config)")

	flag.Parse()

	if showVersion {
		fmt.Printf("%s %skbak%s version %s%s%s\n",
			utils.K8sEmoji, utils.Bold, utils.Reset, utils.Cyan, Version, utils.Reset)
		os.Exit(0)
	}

	if allNamespaces && namespace != "" {
		fmt.Printf("%s %s%sWarning: --namespace flag is ignored when --all-namespaces is used%s\n",
			utils.WarningEmoji, utils.Yellow, utils.Bold, utils.Reset)
		namespace = ""
	}

	// Fail before touching the cluster: an unusable key must never fall back to plain text
	var recipients []age.Recipient
	if recipientsFile != "" {
		var err error
		if recipients, err = utils.LoadRecipients(recipientsFile); err != nil {
			fmt.Printf("%s %s%sError loading age recipients: %v%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, err, utils.Reset)
			os.Exit(1)
		}
	}

	// Initialize Kubernetes client first to validate connectivity
	k8sClient, err := client.NewClient(kubeconfig, verbose)
	if err != nil {
		fmt.Printf("%s %s%sError initializing Kubernetes client: %v%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, err, utils.Reset)
		os.Exit(1)
	}

	// Sortable, unique per second, and free of ':' (invalid in Windows paths)
	backupDir := filepath.Join(outputDir, time.Now().Format("2006-01-02_15-04-05"))
	var namespaces []string
	target := "all namespaces"

	if allNamespaces {
		list, err := k8sClient.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("%s %s%sError listing namespaces: %v%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, err, utils.Reset)
			os.Exit(1)
		}
		for _, ns := range list.Items {
			namespaces = append(namespaces, ns.Name)
		}
		backupDir = filepath.Join(backupDir, "all-namespaces")
	} else {
		// If namespace is not specified, use the kubeconfig (or in-cluster) namespace
		if namespace == "" {
			namespace = k8sClient.Namespace
			if verbose {
				fmt.Printf(" %sUsing current namespace: %s%s\n",
					utils.Cyan, namespace, utils.Reset)
			}
		}
		// Listing in a missing namespace returns nothing, so a typo would pass as an empty backup.
		// Other errors (e.g. no permission to get namespaces) are left to the list calls.
		if _, err := k8sClient.Clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); apierrors.IsNotFound(err) {
			fmt.Printf("%s %s%sNamespace '%s' does not exist%s\n",
				utils.ErrorEmoji, utils.Red, utils.Bold, namespace, utils.Reset)
			os.Exit(1)
		}
		namespaces = []string{namespace}
		target = fmt.Sprintf("namespace '%s'", namespace)
		backupDir = filepath.Join(backupDir, namespace)
	}

	selectedTypes := buildResourceTypeMap(resFlags)
	scope := "all resource types"
	if len(selectedTypes) > 0 {
		scope = "selected resource types"
	}
	if recipients != nil {
		fmt.Printf("%s %sEncrypting with age to %d recipient(s) from '%s'%s\n",
			utils.InfoEmoji, utils.Cyan, len(recipients), recipientsFile, utils.Reset)
	}
	fmt.Printf("%s %s%sStarting backup of %s from %s to '%s'%s\n\n",
		utils.StartEmoji, utils.Blue, utils.Bold, scope, target, backupDir, utils.Reset)

	totalResourceCount, totalErrorCount := 0, 0
	for _, ns := range namespaces {
		nsBackupDir := backupDir
		if allNamespaces {
			nsBackupDir = filepath.Join(backupDir, ns)
			fmt.Printf("%sProcessing namespace: %s%s\n", utils.Blue, ns, utils.Reset)
		}
		resourceCount, errorCount := backup.PerformBackup(k8sClient, ns, nsBackupDir, selectedTypes, recipients, verbose)
		totalResourceCount += resourceCount
		totalErrorCount += errorCount
	}

	switch {
	case totalErrorCount > 0:
		fmt.Printf("\n%s %s%sBackup to %s finished with %d errors (%d resources saved)%s\n",
			utils.ErrorEmoji, utils.Red, utils.Bold, backupDir, totalErrorCount, totalResourceCount, utils.Reset)
		os.Exit(1)
	case totalResourceCount == 0:
		fmt.Printf("\n%s %s%sNo resources found to backup in %s%s\n",
			utils.WarningEmoji, utils.Yellow, utils.Bold, target, utils.Reset)
	default:
		fmt.Printf("\n%s %s%sBackup completed successfully to %s (%d resources total)%s\n",
			utils.SuccessEmoji, utils.Green, utils.Bold, backupDir, totalResourceCount, utils.Reset)
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
		flags.cronjob || flags.job || flags.route

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
	if flags.route {
		selectedTypes["route"] = true
	}

	return selectedTypes
}

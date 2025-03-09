![Project Logo](/assets/logo.jpg)
# kbak - Kubernetes Manifest Backup Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/rogosprojects/kbak?)](https://goreportcard.com/report/github.com/rogosprojects/kbak)
[![GitHub release](https://img.shields.io/github/release/rogosprojects/kbak.svg)](https://github.com/rogosprojects/kbak/releases/latest)

A Go application that backs up Kubernetes resources from a specified namespace by exporting YAML manifests, organized by resource kind.

## Use Cases

- **Disaster Recovery**: Create point-in-time snapshots of your Kubernetes resources for quick restoration in case of accidental deletions or cluster failures.
- **Environment Migration**: Export manifests from one environment (e.g., staging) for deployment in another environment (e.g., production) with necessary modifications.
- **Version Control**: Store your Kubernetes configurations in version control to track changes and maintain configuration history over time.
- **Auditing and Compliance**: Generate snapshots of your cluster state for auditing purposes and compliance requirements.

## Features

- Exports all standard Kubernetes resources from a namespace or all namespaces
- Uses the current namespace from kubeconfig when no namespace is specified
- Organizes backups by resource kind in separate directories
- Thoroughly cleans manifests by removing server-side and cluster-specific fields
- Timestamp-based backup directories
- Colorful and descriptive console output with emojis
- Resource type filtering for selective backups

## Installation

### From binaries

Simply download [latest binaries](https://github.com/rogosprojects/kbak/releases/latest).

### Building from source
```bash
go install github.com/rogosprojects/kbak/cmd/kbak@latest
```
or

```bash
# Clone the repository
git clone https://github.com/rogosprojects/kbak.git && cd kbak

# Build the application
go build -o kbak ./cmd/kbak

# Optional: install system-wide (may require sudo)
sudo cp kbak /usr/local/bin/
```
#### Using Docker

```bash
# Build the Docker image
docker build -t kbak:latest .

# Run with your kubeconfig mounted
docker run --rm -v ~/.kube:/root/.kube -v $(pwd)/backups:/backups kbak:latest --namespace your-namespace
```

## Usage

```bash
# Show version information
./kbak --version

# Backup resources from your current namespace (from kubeconfig)
./kbak

# Backup a specific namespace
./kbak --namespace your-namespace

# Backup all namespaces
./kbak --all-namespaces

# Specify a custom kubeconfig file
./kbak --namespace your-namespace --kubeconfig /path/to/kubeconfig

# Specify a custom output directory
./kbak --namespace your-namespace --output /path/to/backup/dir

# Backup only specific resource types (e.g., only ConfigMaps and Secrets)
./kbak --namespace your-namespace --configmap --secret --all-resources=false

# Backup only pods and deployments
./kbak --namespace your-namespace --pod --deployment --all-resources=false

# Backup resources with verbose output
./kbak --namespace your-namespace --verbose
```

### Resource Type Filtering

You can selectively backup specific resource types using these flags:

```
--all-resources   Backup all resource types (default: true)
--pod             Backup only pods
--deployment      Backup only deployments
--service         Backup only services
--configmap       Backup only configmaps
--secret          Backup only secrets
--pvc             Backup only persistent volume claims
--serviceaccount  Backup only service accounts
--statefulset     Backup only statefulsets
--daemonset       Backup only daemonsets
--ingress         Backup only ingresses
--role            Backup only roles
--rolebinding     Backup only rolebindings
--cronjob         Backup only cronjobs
--job             Backup only jobs
```

When using resource type flags, set `--all-resources=false` to backup only the specified types.


## Supported Resources

The tool automatically backs up the following resource types:

- Core resources: Pods, Services, ConfigMaps, Secrets, PersistentVolumeClaims, ServiceAccounts
- Apps resources: Deployments, StatefulSets, DaemonSets
- Networking resources: Ingresses
- Batch resources: Jobs, CronJobs
- RBAC resources: Roles, RoleBindings

## Output Structure

### Single Namespace Backup
```
02Jan2006-15:04/
└── namespace/
    ├── Pod/
    │   ├── my-pod.yaml
    │   └── ...
    ├── Deployment/
    │   ├── my-deployment.yaml
    │   └── ...
    ├── Service/
    │   ├── my-service.yaml
    │   └── ...
    └── ...
```

### All Namespaces Backup
```
02Jan2006-15:04/
└── all-namespaces/
    ├── namespace1/
    │   ├── Pod/
    │   │   ├── my-pod.yaml
    │   │   └── ...
    │   ├── Deployment/
    │   │   └── ...
    │   └── ...
    ├── namespace2/
    │   ├── Pod/
    │   │   └── ...
    │   └── ...
    └── ...
```
## License

MIT License
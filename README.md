# kbak - Kubernetes Manifest Backup Tool

A Go application that backs up Kubernetes resources from a specified namespace by exporting YAML manifests, organized by resource kind.

## Features

- Exports all standard Kubernetes resources from a namespace
- Organizes backups by resource kind in separate directories
- Thoroughly cleans manifests by removing server-side and cluster-specific fields
- Timestamp-based backup directories
- Simple command-line interface

## Usage

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/kbak.git
cd kbak

# Build the application
go build -o kbak .
```

### Running the application

```bash
# Backup a namespace using the current kubeconfig
./kbak --namespace your-namespace

# Specify a custom kubeconfig file
./kbak --namespace your-namespace --kubeconfig /path/to/kubeconfig

# Specify a custom output directory
./kbak --namespace your-namespace --output /path/to/backup/dir
```

### Using Docker

```bash
# Build the Docker image
docker build -t kbak:latest .

# Run with your kubeconfig mounted
docker run --rm -v ~/.kube:/root/.kube -v $(pwd)/backups:/backups kbak:latest --namespace your-namespace
```

## Supported Resources

The tool automatically backs up the following resource types:

- Core resources: Pods, Services, ConfigMaps, Secrets, PersistentVolumeClaims, ServiceAccounts
- Apps resources: Deployments, StatefulSets, DaemonSets
- Networking resources: Ingresses
- Batch resources: Jobs, CronJobs
- RBAC resources: Roles, RoleBindings

## Output Structure

Backups are organized as follows:

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


## License

MIT License
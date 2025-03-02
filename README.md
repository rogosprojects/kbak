![Project Logo](/assets/logo.jpg)
# kbak - Kubernetes Manifest Backup Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/rogosprojects/kbak?)](https://goreportcard.com/report/github.com/rogosprojects/kbak)
[![GitHub release](https://img.shields.io/github/release/rogosprojects/kbak.svg)](https://github.com/rogosprojects/kbak/releases/latest)

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

# Build with a specific version
go build -ldflags="-X main.Version=1.0.0" -o kbak .
```

### Running the application

```bash
# Show version information
./kbak --version

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

# Build with a specific version
docker build --build-arg VERSION=1.0.0 -t kbak:1.0.0 .

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


## CI/CD Integration

When using GitHab Actions, you can automatically set the version based on git tags:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Get Version from Tag
        id: get_version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_ENV
        if: startsWith(github.ref, 'refs/tags/')

      - name: Build
        run: go build -ldflags="-X main.Version=${VERSION:-dev}" -o kbak .

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: ${{ github.event_name != 'pull_request' }}
          tags: yourregistry/kbak:${{ env.VERSION || 'latest' }}
          build-args: |
            VERSION=${{ env.VERSION || 'dev' }}
```

## License

MIT License
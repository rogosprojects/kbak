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
- Leaves out objects the cluster recreates by itself (Pods and Jobs owned by a controller, the `kube-root-ca.crt` ConfigMap)
- Timestamp-based backup directories
- Colorful and descriptive console output with emojis
- Resource type filtering for selective backups
- Optional [age](https://age-encryption.org) encryption of every file to public keys, so the machine running the backup can't read it back

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

# Run with your kubeconfig mounted, as your own user so you can read the files
mkdir -p backups
docker run --rm --user "$(id -u):$(id -g)" \
  -v ~/.kube/config:/kube/config:ro -e KUBECONFIG=/kube/config \
  -v "$(pwd)/backups:/backups" kbak:latest --namespace your-namespace
```

Backup files are created readable only by their owner (`0600`, folders `0700`), because they include Secrets.

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

# Specify a custom kubeconfig file (default: $KUBECONFIG, then ~/.kube/config, then the in-cluster service account)
./kbak --namespace your-namespace --kubeconfig /path/to/kubeconfig

# Specify a custom output directory
./kbak --namespace your-namespace --output /path/to/backup/dir

# Backup only specific resource types (e.g., only ConfigMaps and Secrets)
./kbak --namespace your-namespace --configmap --secret

# Backup only pods and deployments
./kbak --namespace your-namespace --pod --deployment

# Backup resources with verbose output
./kbak --namespace your-namespace --verbose

# Encrypt every file with age (see Encryption)
./kbak --namespace your-namespace --age-recipients-file recipients.txt
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
--route           Backup only OpenShift routes
```

Any resource type flag limits the backup to the selected types. Without one, every type is backed up.

The exit code is `1` if any resource could not be listed or written (for example, missing RBAC permissions) or the namespace does not exist, so scheduled backups can alert on failures.

## Supported Resources

The tool automatically backs up the following resource types:

- Core resources: Pods, Services, ConfigMaps, Secrets, PersistentVolumeClaims, ServiceAccounts
- Apps resources: Deployments, StatefulSets, DaemonSets
- Networking resources: Ingresses
- Batch resources: Jobs, CronJobs
- RBAC resources: Roles, RoleBindings
- OpenShift resources: Routes (skipped on clusters without the `route.openshift.io` API)

## Output Structure

Each run writes to a new folder named after its start time (`YYYY-MM-DD_HH-MM-SS`), so folders sort chronologically.

### Single Namespace Backup
```
2026-01-02_15-04-05/
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
2026-01-02_15-04-05/
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

## Encryption

With `--age-recipients-file`, every manifest is encrypted with [age](https://age-encryption.org) and written as `<name>.yaml.age` instead of `<name>.yaml`. The folder layout stays the same. Plain text never reaches the disk.

kbak only needs **public keys**. The private key stays elsewhere (an admin laptop, a vault), so a stolen backup host or CronJob cannot decrypt the backups.

```bash
# Once, on the machine that will decrypt. Post-quantum keys are recommended for long-lived backups.
age-keygen -pq -o key.txt              # private key: keep it safe and out of the cluster
age-keygen -y key.txt > recipients.txt # public key: give this to kbak

./kbak --namespace your-namespace --age-recipients-file recipients.txt
```

`recipients.txt` uses the `age -R` format: one native age public key per line (`age1pq1...` or `age1...`), with `#` comments and blank lines allowed. Every file is encrypted to all keys in it, and any of their private keys can decrypt it. Post-quantum and classic keys can't be mixed, and SSH keys are not accepted.

> **If you lose every private key, the backups can't be recovered.** Add a second, offline break-glass key to `recipients.txt`.

If the file is missing, empty, or has an invalid key, kbak exits with `1` before contacting the cluster. If an object fails to encrypt, it is counted as an error and nothing is written for it. kbak never falls back to plain text.

In a CronJob, mount the recipients file from a ConfigMap, since it holds no secret:

```yaml
containers:
  - name: kbak
    image: kbak:latest
    args: ["--all-namespaces", "--age-recipients-file", "/etc/kbak/recipients.txt"]
    volumeMounts:
      - { name: recipients, mountPath: /etc/kbak, readOnly: true }
volumes:
  - name: recipients
    configMap: { name: kbak-age-recipients }
```

Decrypt with the [age CLI](https://github.com/FiloSottile/age#installation). kbak is not needed:

```bash
# One object, straight to the cluster
age -d -i key.txt backups/<timestamp>/your-namespace/Secret/my-secret.yaml.age | kubectl apply -f -

# Decrypt a whole backup next to the encrypted files, then restore it as below.
# umask 077 because age creates its output with your default umask (usually world-readable)
(umask 077; find backups/<timestamp> -name '*.yaml.age' -exec sh -c 'age -d -i key.txt -o "${1%.age}" "$1"' _ {} \;)
```

## Restoring

```bash
kubectl create namespace your-namespace
# ServiceAccounts first: pods and token Secrets are rejected or deleted if their account is missing
kubectl apply -R -f backups/<timestamp>/your-namespace/ServiceAccount
kubectl apply -R -f backups/<timestamp>/your-namespace
```

What to expect:

- Pods and Jobs created by controllers are not in the backup; their Deployments, StatefulSets, DaemonSets and CronJobs recreate them.
- A restored Job runs again.
- PersistentVolumeClaims come back empty and are provisioned again: kbak saves manifests, not volume data.
- Service account token Secrets receive a new token from the target cluster.
- Service `nodePort` values, including auto-assigned ones, are kept and can clash with ports already used on another cluster.

## License

MIT License
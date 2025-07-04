# Tekton AppWrapper Controller Deployment

This directory contains the Kubernetes manifests for deploying the Tekton AppWrapper Controller.

## Prerequisites

- Kubernetes cluster (1.24+)
- `kubectl` configured to communicate with your cluster
- `ko` installed (for building and publishing images)
- Tekton Pipeline installed in `tekton-pipelines` namespace
- MultiKueue installed (for AppWrapper support)
- AppWrapper CRD installed (external dependency from project-codeflare/appwrapper)

## File Structure

```
config/
├── 100-deployment.yaml          # Controller deployment
├── 200-rbac.yaml               # ServiceAccount, ClusterRole, and ClusterRoleBinding
├── 300-config.yaml             # Configuration ConfigMap
├── 400-service.yaml            # Service for metrics
└── README.md                   # This file
```

## Quick Start

### Deploy with ko (Recommended)

```bash
# Deploy everything in one command (builds image and applies manifests)
make deploy-ko

# Or manually:
ko apply -f config/
```

### Deploy with pre-built image

```bash
# Apply manifests directly
kubectl apply -f config/
```

## Manual Deployment Steps

1. **Apply RBAC and configuration:**
   ```bash
   kubectl apply -f config/200-rbac.yaml
   kubectl apply -f config/300-config.yaml
   ```

2. **Deploy controller:**
   ```bash
   ko apply -f config/100-deployment.yaml
   ```

3. **Create service:**
   ```bash
   kubectl apply -f config/400-service.yaml
   ```

## Configuration

### Environment Variables

The controller supports the following environment variables:

- `SYSTEM_NAMESPACE`: The namespace where the controller is running
- `POD_NAME`: The name of the pod running the controller
- `METRICS_DOMAIN`: Domain for metrics (default: `tekton.dev/tekton-appwrapper`)

### ConfigMap

The `tekton-appwrapper-config` ConfigMap contains:

- `default-queue`: Default queue name for AppWrappers
- `logging-config`: JSON logging configuration
- `observability-config`: Metrics and observability settings

## Monitoring

### Metrics

The controller exposes Prometheus metrics on port 9090:

```bash
# Port forward to access metrics
kubectl port-forward -n tekton-pipelines svc/tekton-appwrapper-controller 9090:9090

# Access metrics
curl http://localhost:9090/metrics
```

### Logs

```bash
# View controller logs
kubectl logs -n tekton-pipelines -l app=tekton-appwrapper-controller -f
```

## Troubleshooting

### Check controller status

```bash
# Check if pods are running
kubectl get pods -n tekton-pipelines

# Check deployment status
kubectl get deployment -n tekton-pipelines tekton-appwrapper-controller

# Check events
kubectl get events -n tekton-pipelines --sort-by='.lastTimestamp'
```

### Verify RBAC

```bash
# Check if service account has proper permissions
kubectl auth can-i create appwrappers --as=system:serviceaccount:tekton-pipelines:tekton-appwrapper-controller
kubectl auth can-i get pipelineruns --as=system:serviceaccount:tekton-pipelines:tekton-appwrapper-controller
```

## Uninstallation

```bash
# Remove everything
make undeploy-ko

# Or manually:
ko delete -f config/
```

## Development

### Build and test locally

```bash
# Build binary
make build

# Run locally
make run

# Run tests
make test
```

### Build image with ko

```bash
# Build image locally
make ko-build

# Publish image
make ko-publish
```

## External Dependencies

### AppWrapper CRD

The AppWrapper CRD is an external dependency that must be installed separately:

```bash
# Install AppWrapper CRD
kubectl apply -f https://raw.githubusercontent.com/project-codeflare/appwrapper/main/config/crd/bases/workload.codeflare.dev_appwrappers.yaml
```

This CRD is required for the controller to function properly as it defines the AppWrapper custom resource. 
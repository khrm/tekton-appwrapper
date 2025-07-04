# Tekton AppWrapper Controller

A Kubernetes controller that integrates Tekton `PipelineRun` resources with MultiKueue by creating AppWrappers for resource management and scheduling.

## Overview

This controller watches for Tekton `PipelineRun` resources and automatically creates corresponding AppWrappers that can be managed by MultiKueue for advanced scheduling and resource allocation.

## Features

- **Automatic AppWrapper Creation**: Creates AppWrappers for PipelineRuns with resource requirements
- **Resource Management**: Extracts resource requests from PipelineRun annotations
- **MultiKueue Integration**: Supports MultiKueue queue management
- **Knative/pkg Pattern**: Built using Knative's proven controller patterns
- **Production Ready**: Includes RBAC, monitoring, and observability

## Prerequisites

- Kubernetes cluster (v1.24+)
- Tekton Pipeline installed
- MultiKueue installed (optional, for advanced scheduling)
- AppWrapper CRD installed
- `ko` tool installed

### Installing Prerequisites

```bash
# Install ko
go install github.com/ko-build/ko@latest

# Install AppWrapper CRD
kubectl apply -f https://raw.githubusercontent.com/project-codeflare/appwrapper/main/config/crd/bases/workload.codeflare.dev_appwrappers.yaml

# Install Tekton Pipeline (if not already installed)
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
```

## Quick Start

### 1. Build and Deploy

```bash
# Build and deploy in one command
make ko-apply

# Or manually:
ko apply -f config/
```

### 2. Verify Installation

```bash
# Check controller status
make status

# View controller logs
make logs
```

### 3. Test with a PipelineRun

Create a PipelineRun with resource annotations:

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  name: test-pipeline
  annotations:
    kueue.tekton.dev/requests-cpu: "500m"
    kueue.tekton.dev/requests-memory: "1Gi"
spec:
  pipelineSpec:
    tasks:
      - name: echo
        taskSpec:
          steps:
            - name: echo
              image: ubuntu
              script: |
                echo "Hello from Tekton AppWrapper!"
```

The controller will automatically create an AppWrapper for this PipelineRun.

## Development

### Local Development

```bash
# Setup development environment
make dev-setup

# Run tests
make test

# Format code
make fmt

# Lint code
make lint

# Cleanup
make dev-cleanup
```

### Building Locally

```bash
# Build binary
make build

# Build container image
make ko-build

# Build and publish image
make ko-publish
```

## Configuration

### Resource Annotations

The controller looks for the following annotations on PipelineRuns:

- `kueue.tekton.dev/requests-cpu`: CPU request (e.g., "500m", "1")
- `kueue.tekton.dev/requests-memory`: Memory request (e.g., "1Gi", "512Mi")
- `kueue.tekton.dev/requests-storage`: Storage request
- `kueue.tekton.dev/requests-ephemeral-storage`: Ephemeral storage request

### Queue Configuration

AppWrappers are created with the `default-queue` by default. You can modify this in the controller code or make it configurable via environment variables.

## Architecture

### Controller Pattern

This controller follows the Knative/pkg pattern:

- Uses informers for efficient resource watching
- Implements the `Reconcile` interface
- Uses dynamic client for AppWrapper operations
- Follows Knative's injection patterns

### Resource Flow

1. **PipelineRun Created**: Controller detects new PipelineRun
2. **Resource Extraction**: Extracts resource requirements from annotations
3. **AppWrapper Creation**: Creates AppWrapper with PipelineRun embedded
4. **MultiKueue Scheduling**: AppWrapper can be scheduled by MultiKueue
5. **Resource Management**: AppWrapper manages resource allocation

## Monitoring

The controller exposes metrics on port 9090:

```bash
# Port forward to access metrics
kubectl port-forward -n tekton-appwrapper-system svc/tekton-appwrapper-controller 9090:9090

# Access metrics
curl http://localhost:9090/metrics
```

## Troubleshooting

### Common Issues

1. **AppWrapper CRD not found**
   ```bash
   make install-crds
   ```

2. **Controller not starting**
   ```bash
   make logs
   kubectl describe pod -n tekton-appwrapper-system -l app.kubernetes.io/name=tekton-appwrapper
   ```

3. **Permission denied**
   ```bash
   kubectl auth can-i create appwrappers --namespace default
   ```

### Debug Mode

Enable debug logging by updating the config-logging ConfigMap:

```bash
kubectl patch configmap config-logging -n tekton-appwrapper-system --patch '{"data":{"loglevel.controller":"debug"}}'
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make lint`
6. Submit a pull request

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## Support

For issues and questions:

- Create an issue in this repository
- Check the [Tekton community](https://github.com/tektoncd/community) for general Tekton support
- Check the [MultiKueue community](https://github.com/project-codeflare/multikueue) for MultiKueue support 
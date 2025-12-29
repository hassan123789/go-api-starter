# Kubernetes Manifests for go-api-starter

Production-ready Kubernetes manifests using Kustomize for environment-specific configurations.

## Structure

```
k8s/
├── base/                        # Base configurations
│   ├── kustomization.yaml       # Kustomize configuration
│   ├── namespace.yaml           # Namespace definition
│   ├── serviceaccount.yaml      # ServiceAccount
│   ├── configmap.yaml           # Non-sensitive configuration
│   ├── secret.yaml              # Sensitive data (example only!)
│   ├── deployment.yaml          # Main application deployment
│   ├── service.yaml             # ClusterIP service
│   ├── hpa.yaml                 # Horizontal Pod Autoscaler
│   ├── pdb.yaml                 # Pod Disruption Budget
│   ├── ingress.yaml             # Ingress configuration
│   └── networkpolicy.yaml       # Network policies
└── overlays/
    ├── development/             # Dev environment overrides
    │   ├── kustomization.yaml
    │   └── deployment-patch.yaml
    └── production/              # Prod environment overrides
        ├── kustomization.yaml
        ├── deployment-patch.yaml
        ├── hpa-patch.yaml
        └── ingress-patch.yaml
```

## Usage

### Prerequisites

- kubectl configured with cluster access
- Kustomize (built into kubectl 1.14+)

### Deploy to Development

```bash
# Preview
kubectl kustomize deploy/k8s/overlays/development

# Apply
kubectl apply -k deploy/k8s/overlays/development
```

### Deploy to Production

```bash
# Preview
kubectl kustomize deploy/k8s/overlays/production

# Apply
kubectl apply -k deploy/k8s/overlays/production
```

### Build Docker Image

```bash
# Build and push to registry
docker build -t ghcr.io/hassan123789/go-api-starter:v1.0.0 .
docker push ghcr.io/hassan123789/go-api-starter:v1.0.0

# Update kustomization.yaml with new tag
cd deploy/k8s/overlays/production
kustomize edit set image go-api-starter=ghcr.io/hassan123789/go-api-starter:v1.0.0
```

## Security Features

| Feature | Description |
|---------|-------------|
| **Non-root user** | Runs as UID 65532 (distroless nonroot) |
| **Read-only filesystem** | Container filesystem is read-only |
| **No privilege escalation** | `allowPrivilegeEscalation: false` |
| **Dropped capabilities** | All Linux capabilities dropped |
| **Network policies** | Ingress/Egress traffic restricted |
| **Pod anti-affinity** | Pods spread across nodes |
| **Seccomp profile** | RuntimeDefault seccomp enabled |

## Health Checks

| Probe | Path | Purpose |
|-------|------|---------|
| **Startup** | `/health` | Wait for app initialization |
| **Liveness** | `/live` | Restart if unhealthy |
| **Readiness** | `/ready` | Remove from service if not ready |

## Scaling

The HPA automatically scales based on:

- CPU utilization (target: 70%)
- Memory utilization (target: 80%)

```yaml
# Development: 1 replica
# Production: 3-20 replicas
```

## Secrets Management

⚠️ **WARNING**: The `secret.yaml` contains placeholder values only!

For production, use one of:

- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [External Secrets Operator](https://external-secrets.io/)
- [HashiCorp Vault](https://www.vaultproject.io/)
- [SOPS](https://github.com/getsops/sops)

## Local Testing with Kind/Minikube

```bash
# Create cluster
kind create cluster --name go-api-starter

# Load local image
kind load docker-image go-api-starter:latest --name go-api-starter

# Deploy
kubectl apply -k deploy/k8s/overlays/development

# Port forward
kubectl port-forward -n go-api-starter-dev svc/dev-go-api-starter 8080:80

# Test
curl http://localhost:8080/health
```

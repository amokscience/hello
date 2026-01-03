# Multi-Environment Deployment Guide

This application is configured to deploy to multiple environments using Kustomize and ArgoCD.

## Environments

- **dev** - Development environment (1 replica, minimal resources)
- **qa** - Quality Assurance environment (2 replicas, moderate resources)
- **staging** - Staging environment (2 replicas, moderate resources)
- **prod** - Production environment (3 replicas, higher resources)

Each environment has:
- Its own namespace (`hello-dev`, `hello-qa`, `hello-staging`, `hello-prod`)
- Environment-specific ENVIRONMENT variable
- Scaled replicas and resources appropriate for the environment
- Name prefix (e.g., `dev-hello-app`, `prod-hello-app`)

## Local Development

Run locally with environment variables:

```bash
# Development
$env:HELLO="Hello, World!"
$env:ENVIRONMENT="dev"
go run main.go
# Access: http://localhost:8080
```

## Kustomize Deployment

Deploy using Kustomize to any environment:

```bash
# Development
kubectl apply -k kustomize/overlays/dev

# QA
kubectl apply -k kustomize/overlays/qa

# Staging
kubectl apply -k kustomize/overlays/staging

# Production
kubectl apply -k kustomize/overlays/prod
```

### Preview the generated manifests:

```bash
# See what will be deployed for dev
kustomize build kustomize/overlays/dev

# See what will be deployed for prod
kustomize build kustomize/overlays/prod
```

## ArgoCD Deployment

### Single Application per Environment

Deploy individual environments:

```bash
# Dev
kubectl apply -f argocd-apps-multienv.yaml --select='name=hello-app-dev'

# QA
kubectl apply -f argocd-apps-multienv.yaml --select='name=hello-app-qa'

# Staging
kubectl apply -f argocd-apps-multienv.yaml --select='name=hello-app-staging'

# Production
kubectl apply -f argocd-apps-multienv.yaml --select='name=hello-app-prod'
```

### Deploy All Environments:

```bash
kubectl apply -f argocd-apps-multienv.yaml
```

This will create 4 ArgoCD Applications, one for each environment.

**Note:** Dev, QA, and Staging have automated sync enabled. Production has manual sync for safety.

## Environment Variables by Environment

| Environment | ENVIRONMENT Value | Replicas | Memory Request | Memory Limit |
|-------------|-------------------|----------|----------------|--------------|
| dev         | dev               | 1        | 64Mi           | 128Mi        |
| qa          | qa                | 2        | 128Mi          | 256Mi        |
| staging     | staging           | 2        | 128Mi          | 256Mi        |
| prod        | prod              | 3        | 256Mi          | 512Mi        |

## Message Format

The application displays: `Hello, World! - <environment>`

Example outputs:
- Local: `Hello, World! - dev`
- QA: `Hello, World! - qa`
- Production: `Hello, World! - prod`

## Project Structure

```
kustomize/
├── base/                          # Common configuration
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/                      # Environment-specific overrides
    ├── dev/
    │   └── kustomization.yaml
    ├── qa/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        └── kustomization.yaml
```

## Customization

To customize an environment:

1. Edit the corresponding `kustomize/overlays/<env>/kustomization.yaml`
2. Modify replicas, environment variables, or resource limits
3. Commit and push to git
4. ArgoCD will automatically sync the changes (if auto-sync is enabled)

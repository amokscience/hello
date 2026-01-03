# Docker Hub Setup

This project uses GitHub Actions to automatically build and push Docker images to Docker Hub.

## Prerequisites

1. **Docker Hub Account**
   - Create one at https://hub.docker.com if you don't have one
   - Note your username

2. **Docker Hub Access Token**
   - Go to https://hub.docker.com/settings/security
   - Click "New Access Token"
   - Give it a name like "github-actions"
   - Set permissions for "Read, Write, Delete"
   - Copy the token (you'll need it shortly)

## GitHub Secrets Setup

Add the following secrets to your GitHub repository:

1. Go to: `https://github.com/amokscience/hello/settings/secrets/actions`

2. Click **New repository secret** and add:

   | Secret Name | Value |
   |-------------|-------|
   | `DOCKER_USERNAME` | `amokscience` |
   | `DOCKER_PASSWORD` | Your Docker Hub access token (not your password!) |

## How It Works

The workflow (`.github/workflows/docker-build-push.yaml`) automatically:

1. **Triggers on:**
   - Push to `main` branch
   - Changes to source files, Dockerfile, or workflow file
   - Manual trigger via GitHub Actions UI

2. **Builds:**
   - Uses Docker Buildx for multi-platform builds
   - Caches layers for faster subsequent builds
   - Tags with:
     - `latest`
     - Git branch name
     - Git commit SHA
     - Semantic version tags (if you use git tags like `v1.0.0`)

3. **Pushes to Docker Hub:**
   - Repository: `amokscience/hello`
   - Available at: `https://hub.docker.com/r/amokscience/hello`

## Image Tags

The workflow automatically creates multiple tags:

```
docker.io/amokscience/hello:latest
docker.io/amokscience/hello:main
docker.io/amokscience/hello:sha-abc1234
docker.io/amokscience/hello:v1.0.0  (if you tag releases)
```

## Manual Trigger

To manually trigger a build without pushing code:

1. Go to: `https://github.com/amokscience/hello/actions`
2. Select "Build and Push Docker Image" workflow
3. Click "Run workflow"
4. Select the branch and click "Run workflow"

## Using the Image

After a successful build, use the image in Kubernetes:

```bash
# Pull the latest image
docker pull amokscience/hello:latest

# Or run it locally
docker run -e HELLO="Hello, World!" \
           -e ENVIRONMENT="dev" \
           -p 8080:8080 \
           amokscience/hello:latest
```

## Update Kustomize Overlays

Update your Kustomize overlays to use your Docker Hub image:

In `kustomize/base/deployment.yaml`, the image is already set to:

```yaml
containers:
- name: hello
  image: amokscience/hello:latest
  imagePullPolicy: Always
```

Or use a specific tag:

```yaml
image: amokscience/hello:v1.0.0
```

## Troubleshooting

**Build fails with auth error:**
- Verify secrets are set correctly in GitHub
- Don't include `docker.io/` prefix in `DOCKER_USERNAME`
- Use an access token, not your Docker Hub password

**Image not found after push:**
- Wait a moment for Docker Hub to process the push
- Check your Docker Hub repository at https://hub.docker.com/r/amokscience/hello

**Want to disable auto-push:**
- Comment out or remove the `push: true` line in the workflow
- Remove the `docker/login-action` step
- The image will still build locally in the runner

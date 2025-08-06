---
title: "Docker Contributions"
sidebar_label: "Docker"
---

# Contributing to Nixopus Docker Builds

This guide provides detailed instructions for contributing to Nixopus Docker builds and container optimization.

## Overview

Docker is central to Nixopus deployment strategy. Contributions to Docker builds can include:

- Optimizing Docker images for size and security
- Improving build performance
- Enhancing multi-stage builds
- Configuring container orchestration
- Supporting new container platforms

## Understanding the Docker Structure

Nixopus uses Docker for containerization with the following key components:

```
/
├── api/
│   └── Dockerfile         # API service Dockerfile
├── view/
│   └── Dockerfile         # Frontend Dockerfile
├── docker-compose.yml     # Main compose file
└── docker-compose-staging.yml  # Staging environment compose file
```

## Best Practices for Docker Contributions

### 1. Image Optimization

When optimizing Docker images:

1. **Use Multi-Stage Builds**

   Example improvement for the API Dockerfile:

   ```dockerfile
   # Build stage
   FROM golang:1.21-alpine AS builder
   
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   
   COPY . .
   RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
   
   # Final stage
   FROM alpine:3.18
   
   RUN apk --no-cache add ca-certificates tzdata
   
   WORKDIR /app
   COPY --from=builder /app/app .
   COPY --from=builder /app/migrations ./migrations
   
   # Add non-root user
   RUN adduser -D -g '' appuser
   USER appuser
   
   CMD ["./app"]
   ```

2. **Minimize Layer Size**

   Best practices:
   - Combine RUN commands with `&&`
   - Clean up in the same layer
   - Use `.dockerignore` to exclude unnecessary files

   Example:

   ```dockerfile
   RUN apk --no-cache add \
       curl \
       tzdata \
       ca-certificates \
       openssl \
       && rm -rf /var/cache/apk/*
   ```

   Example `.dockerignore`:

   ```
   .git
   .github
   .vscode
   node_modules
   npm-debug.log
   Dockerfile*
   docker-compose*
   *.md
   test/
   **/__tests__
   coverage/
   ```

3. **Use Specific Versions**

   Bad:

   ```dockerfile
   FROM node:latest
   ```

   Good:

   ```dockerfile
   FROM node:20.6.1-alpine3.18
   ```

4. **Leverage BuildKit Features**

   ```dockerfile
   # syntax=docker/dockerfile:1.4
   
   # Enable BuildKit cache mounts for faster builds
   RUN --mount=type=cache,target=/var/cache/apt \
       --mount=type=cache,target=/var/lib/apt \
       apt-get update && apt-get install -y --no-install-recommends \
       curl \
       ca-certificates
   ```

### 2. Security Improvements

1. **Use Non-Root Users**

   ```dockerfile
   # Create a non-root user
   RUN addgroup -S appgroup && adduser -S appuser -G appgroup
   
   # Set permissions
   COPY --chown=appuser:appgroup . .
   
   # Switch to non-root user
   USER appuser
   ```

2. **Scan Images for Vulnerabilities**

   Implement scanning in your local workflow and document it:

   ```bash
   # Using Docker Scout
   docker scout cves nixopus-api:latest
   
   # Using Trivy
   trivy image nixopus-api:latest
   ```

3. **Minimal Base Images**

   Replace general images with minimal alternatives:

   - Use `alpine` instead of full `debian`
   - Use `distroless` images for production

   Example:

   ```dockerfile
   # Final production stage
   FROM gcr.io/distroless/static-debian11
   
   COPY --from=builder /app/app /
   
   CMD ["/app"]
   ```

### 3. Docker Compose Improvements

1. **Environment Management**

   ```yaml
   services:
     api:
       env_file: 
         - .env.common
         - .env.${NIXOPUS_ENV:-production}
   ```

2. **Healthcheck Enhancements**

   ```yaml
   services:
     api:
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:8443/health"]
         interval: 10s
         timeout: 5s
         retries: 3
         start_period: 30s
   ```

3. **Resource Constraints**

   ```yaml
   services:
     api:
       deploy:
         resources:
           limits:
             cpus: '0.5'
             memory: 512M
           reservations:
             cpus: '0.1'
             memory: 128M
   ```

4. **Dependency Management**

   ```yaml
   services:
     api:
       depends_on:
         db:
           condition: service_healthy
         redis:
           condition: service_healthy
   ```

### 4. CI/CD Integration

1. **Automated Builds**

   Example GitHub Actions workflow for Docker builds:

   ```yaml
   name: Docker Image CI
   
   on:
     push:
       branches: [ main ]
       tags: [ 'v*' ]
     pull_request:
       branches: [ main ]
   
   jobs:
     build:
       runs-on: ubuntu-latest
       strategy:
         matrix:
           service: [api, view]
       steps:
       - uses: actions/checkout@v3
       
       - name: Set up Docker Buildx
         uses: docker/setup-buildx-action@v2
       
       - name: Login to GitHub Container Registry
         if: github.event_name != 'pull_request'
         uses: docker/login-action@v2
         with:
           registry: ghcr.io
           username: ${{ github.actor }}
           password: ${{ secrets.GITHUB_TOKEN }}
       
       - name: Extract metadata
         id: meta
         uses: docker/metadata-action@v4
         with:
           images: ghcr.io/${{ github.repository }}-${{ matrix.service }}
           tags: |
             type=semver,pattern={{version}}
             type=semver,pattern={{major}}.{{minor}}
             type=semver,pattern={{major}}
             type=ref,event=branch
             type=sha
       
       - name: Build and push
         uses: docker/build-push-action@v4
         with:
           context: ./${{ matrix.service }}
           push: ${{ github.event_name != 'pull_request' }}
           tags: ${{ steps.meta.outputs.tags }}
           labels: ${{ steps.meta.outputs.labels }}
           cache-from: type=gha,scope=${{ matrix.service }}
           cache-to: type=gha,mode=max,scope=${{ matrix.service }}
   ```

2. **Image Testing**

   Add automated testing for Docker images:

   ```yaml
   - name: Test image
     run: |
       docker run --rm ${{ steps.meta.outputs.tags }} health-check
   ```

### 5. Container Orchestration

1. **Kubernetes Support**

   Create Kubernetes manifests for Nixopus:

   ```yaml
   # kubernetes/api-deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: nixopus-api
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: nixopus-api
     template:
       metadata:
         labels:
           app: nixopus-api
       spec:
         containers:
         - name: api
           image: ghcr.io/raghavyuva/nixopus-api:latest
           ports:
           - containerPort: 8443
           env:
           - name: DB_HOST
             valueFrom:
               configMapKeyRef:
                 name: nixopus-config
                 key: db-host
           resources:
             limits:
               cpu: "500m"
               memory: "512Mi"
             requests:
               cpu: "100m"
               memory: "128Mi"
           livenessProbe:
             httpGet:
               path: /health
               port: 8443
             initialDelaySeconds: 30
             periodSeconds: 10
           readinessProbe:
             httpGet:
               path: /health
               port: 8443
             initialDelaySeconds: 5
             periodSeconds: 5
   ```

2. **Docker Swarm Support**

   Create Docker Swarm deployment examples:

   ```yaml
   version: '3.8'
   
   services:
     api:
       image: ghcr.io/raghavyuva/nixopus-api:latest
       deploy:
         replicas: 2
         update_config:
           parallelism: 1
           delay: 10s
           order: start-first
         restart_policy:
           condition: on-failure
           max_attempts: 3
           window: 120s
         resources:
           limits:
             cpus: '0.5'
             memory: 512M
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:8443/health"]
         interval: 10s
         timeout: 5s
         retries: 3
         start_period: 30s
       networks:
         - nixopus-network
   
   networks:
     nixopus-network:
       driver: overlay
       attachable: true
   ```

## Advanced Docker Features

### 1. BuildKit Features

1. **Mount Secrets in Build**

   ```dockerfile
   # syntax=docker/dockerfile:1.4
   
   FROM alpine
   
   RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
       npm ci --production
   ```

   Usage:

   ```bash
   docker build --secret id=npmrc,src=.npmrc -t nixopus-view .
   ```

2. **Leverage Build Cache**

   ```dockerfile
   # Cache node modules
   COPY package.json yarn.lock ./
   RUN --mount=type=cache,target=/root/.yarn \
       yarn install --frozen-lockfile
   ```

### 2. Image Variants

Create specialized image variants:

1. **Development Image**

   ```dockerfile
   FROM nixopus-api:latest AS production
   
   FROM nixopus-api-base:latest AS development
   RUN apk add --no-cache curl jq vim
   COPY --from=production /app /app
   
   # Add development tools and configurations
   COPY tools/dev.sh /usr/local/bin/dev-tools
   RUN chmod +x /usr/local/bin/dev-tools
   
   CMD ["sh", "-c", "dev-tools && ./app"]
   ```

2. **Testing Image**

   ```dockerfile
   FROM nixopus-api-base:latest AS testing
   
   COPY . .
   RUN go test -v ./...
   
   # If tests pass, build the app
   RUN go build -o app .
   
   # Minimal test runner image
   FROM alpine:3.18
   COPY --from=testing /app/app /app
   CMD ["/app", "--run-tests"]
   ```

## Monitoring and Observability

1. **Add Prometheus Metrics**

   ```dockerfile
   # Install Prometheus client
   RUN go get github.com/prometheus/client_golang/prometheus
   
   # Expose metrics endpoint
   EXPOSE 8080
   
   # Add health/readiness checks
   HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
     CMD curl -f http://localhost:8080/health || exit 1
   ```

2. **Log Management**

   ```dockerfile
   # Configure structured logging
   ENV LOG_FORMAT=json
   ENV LOG_LEVEL=info
   
   # Forward logs to stdout/stderr for container orchestration
   RUN ln -sf /dev/stdout /app/logs/app.log && \
       ln -sf /dev/stderr /app/logs/error.log
   ```

## Testing Docker Changes

1. **Local Testing Workflow**

   ```bash
   # Build images
   docker-compose build
   
   # Run tests
   docker-compose run --rm api go test ./...
   
   # Start services
   docker-compose up -d
   
   # Check logs
   docker-compose logs -f
   
   # Verify health
   curl http://localhost:8443/health
   ```

2. **Performance Testing**

   ```bash
   # Check image size
   docker images nixopus-api
   
   # Check startup time
   time docker-compose up -d api
   
   # Check resource usage
   docker stats nixopus-api
   ```

3. **Security Scanning**

   ```bash
   # Scan for vulnerabilities
   docker scout cves nixopus-api:latest
   
   # Check for exposed secrets
   grype nixopus-api:latest
   ```

## Submitting Docker Improvements

1. **Document Changes**

   Include before/after metrics:
   - Image size reduction
   - Build time improvements
   - Security posture enhancements
   - Resource utilization changes

2. **Update Documentation**

   - Update README.md with new Docker features
   - Add examples of new Docker configurations
   - Document any breaking changes

3. **Create Pull Request**

   ```bash
   git add .
   git commit -m "feat(docker): optimize API container size and security"
   git push origin feature/docker-improvements
   ```

4. **PR Details**

   Include in your PR:
   - Purpose of changes
   - Performance metrics
   - Testing methodology
   - Screenshots of before/after

## Common Docker Pitfalls

1. **Image Bloat Issues**
   - Including development dependencies
   - Not cleaning up package manager caches
   - Copying unnecessary files

2. **Security Problems**
   - Running as root
   - Exposing unnecessary ports
   - Using outdated base images
   - Embedding secrets in the image

3. **Performance Issues**
   - Inefficient layer caching
   - Sub-optimal build order
   - Missing health checks
   - Improper resource constraints

## Need Help?

If you need assistance with Docker contributions:

- Join the #docker channel on Discord
- Review the Docker documentation
- Check existing Docker issues for similar improvements

Thank you for improving Nixopus Docker builds!

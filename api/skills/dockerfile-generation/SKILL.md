---
name: dockerfile-generation
description: Generate production-ready multi-stage Dockerfiles per ecosystem with best practices. Use when the user needs a Dockerfile, asks about containerization, or when no Dockerfile exists in the repository.
metadata:
  version: "1.0"
---

# Dockerfile Generation

Generate production Dockerfiles using multi-stage builds. Always optimize for small image size, layer caching, and security.

## General rules

- Always use multi-stage builds (builder + runtime)
- Pin base image versions (e.g. `node:20-alpine`, not `node:latest`)
- Use Alpine or slim variants for runtime
- Copy dependency files first, install, then copy source (layer caching)
- Run as non-root user in production
- Set `NODE_ENV=production` or equivalent
- Include `EXPOSE` directive for the detected port
- Use `.dockerignore` to exclude `node_modules`, `.git`, `dist`, `__pycache__`

## Node.js

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /app/dist ./dist
USER app
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

Variations:
- **Next.js**: Use `output: 'standalone'` in `next.config.js`. Copy `.next/standalone` and `.next/static`.
- **Vite/CRA (static)**: Build in Node, serve with `nginx:alpine`. Copy `dist/` or `build/` to `/usr/share/nginx/html`.
- **pnpm**: Replace `npm ci` with `corepack enable && pnpm install --frozen-lockfile`.
- **yarn**: Replace `npm ci` with `yarn install --frozen-lockfile`.
- **Bun**: Use `oven/bun:1-alpine` as base.

## Go

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/server /server
USER app
EXPOSE 8080
CMD ["/server"]
```

Use `scratch` instead of `alpine` if the binary has no external dependencies and doesn't need a shell.

## Python

```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.12-slim
WORKDIR /app
RUN useradd -r -s /bin/false app
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin
COPY . .
USER app
EXPOSE 8000
CMD ["gunicorn", "app:app", "--bind", "0.0.0.0:8000"]
```

Variations:
- **FastAPI**: `CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]`
- **Django**: `CMD ["gunicorn", "project.wsgi:application", "--bind", "0.0.0.0:8000"]`
- **Poetry**: Copy `pyproject.toml` and `poetry.lock`, use `poetry install --no-dev`.

## Rust

```dockerfile
FROM rust:1.77-alpine AS builder
WORKDIR /app
RUN apk add musl-dev
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
COPY src ./src
RUN touch src/main.rs && cargo build --release

FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/target/release/<binary_name> /app
USER app
EXPOSE 8080
CMD ["/app"]
```

## Java (Maven)

```dockerfile
FROM maven:3.9-eclipse-temurin-21 AS builder
WORKDIR /app
COPY pom.xml ./
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/target/*.jar app.jar
USER app
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]
```

## Static sites (nginx)

For any frontend that produces a `dist/` or `build/` directory:

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Checklist before generating

1. Read the manifest file to confirm ecosystem and framework
2. Check for existing `.dockerignore` — if missing, generate one
3. Detect the actual build output directory (may not be `dist/`)
4. Detect the actual start command from `scripts.start` or framework defaults
5. Detect the correct port from code or config
6. Check if the app needs build-time env vars (e.g. `NEXT_PUBLIC_*`)

## Related Skills

- **`deployment-analysis`** — Run first to determine the ecosystem, framework, port, and build commands before generating a Dockerfile
- **`env-detection`** — Identify build-time vs runtime env vars that need `ARG`/`ENV` directives in the Dockerfile
- **`compose-setup`** — If the app needs a database or other services alongside the Dockerfile

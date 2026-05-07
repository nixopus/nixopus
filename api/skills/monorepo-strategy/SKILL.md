---
name: monorepo-strategy
description: Deploy multi-service monorepo applications — service discovery, dependency ordering, shared build contexts, selective deployment, and compose generation. Use when the repository contains multiple deployable services, apps, or packages.
metadata:
  version: "1.0"
---

# Monorepo Deployment Strategy

## Detection

A repository is a monorepo if any of these are present:

| Signal | Type |
|---|---|
| `apps/` or `services/` directory with multiple subdirectories | Conventional structure |
| `packages/` with shared libraries | Shared code |
| `workspaces` in root `package.json` | npm/yarn/bun workspaces |
| `pnpm-workspace.yaml` | pnpm workspaces |
| `turbo.json` | Turborepo |
| `nx.json` | Nx |
| `lerna.json` | Lerna |
| `rush.json` | Rush |
| `go.work` | Go workspaces |
| Multiple `Dockerfile` files in subdirectories | Multi-service |
| Multiple `Cargo.toml` with `[workspace]` in root | Rust workspace |

## Service Discovery

### Node.js monorepos

1. Read workspace configuration:
   - npm/yarn/bun: `workspaces` array in root `package.json`
   - pnpm: `packages` list in `pnpm-workspace.yaml`
2. For each workspace package, read its `package.json`
3. A package is a **deployable service** if it has a `start` script or a `main`/`module` entry
4. A package is a **shared library** if other packages depend on it but it has no start script

### Go monorepos

1. Read `go.work` for module list
2. Each module with a `main` package (`main.go` or `cmd/` directory) is deployable

### Rust monorepos

1. Read root `Cargo.toml` → `[workspace].members`
2. Each member with `[[bin]]` target or `src/main.rs` is deployable

### Generic (no workspace tool)

1. Scan `apps/`, `services/`, `packages/` for subdirectories
2. Each subdirectory with its own manifest file (`package.json`, `go.mod`, `Cargo.toml`, etc.) is a potential service
3. Each subdirectory with its own `Dockerfile` is a deployable service

## Dependency Graph

Build the service dependency graph before deploying:

1. For each service, read its dependencies on other workspace packages
2. Topologically sort: deploy dependencies before dependents
3. Shared libraries are built first (they're dependencies of services)

### Node.js dependency detection

In `package.json`, workspace dependencies use:
- `"@scope/package": "workspace:*"` (pnpm)
- `"@scope/package": "*"` with the package in `workspaces` (npm/yarn)

### Go dependency detection

In each module's `go.mod`, `require` directives pointing to other workspace modules (matched by module path from `go.work`).

## Build Context

The Docker build context for monorepo services should usually be the **repository root**, not the service subdirectory:

```yaml
services:
  api:
    build:
      context: .                    # repo root
      dockerfile: apps/api/Dockerfile  # service-specific Dockerfile
```

This ensures shared packages, root lockfiles, and workspace configuration are available during build.

### Dockerfile for monorepo service

```dockerfile
FROM node:22-slim AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/api/package.json ./apps/api/
COPY packages/shared/package.json ./packages/shared/
RUN corepack enable && pnpm install --frozen-lockfile

FROM deps AS build
COPY packages/shared ./packages/shared
COPY apps/api ./apps/api
RUN pnpm --filter api build

FROM node:22-slim
WORKDIR /app
COPY --from=build /app/apps/api/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/apps/api/package.json ./
EXPOSE 3001
CMD ["node", "dist/index.js"]
```

Key: copy ALL workspace `package.json` files in the deps stage so the lockfile resolves correctly.

## Selective Deployment

Not every push requires deploying every service:

1. Determine which files changed (from deployment context or git diff)
2. Map changed files to affected services:
   - `apps/api/**` → deploy `api`
   - `packages/shared/**` → deploy ALL services that depend on `shared`
   - Root config (`package.json`, lockfile, `tsconfig.json`) → deploy ALL services
3. Only deploy affected services

## Compose for Monorepos

Generate `docker-compose.yml` with one service per deployable app:

```yaml
services:
  api:
    build:
      context: .
      dockerfile: apps/api/Dockerfile
    ports:
      - "3001:3001"
    depends_on:
      db:
        condition: service_healthy

  web:
    build:
      context: .
      dockerfile: apps/web/Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      - api

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

## Turborepo / Nx Considerations

If `turbo.json` or `nx.json` is present:

- **Turbo**: Use `turbo run build --filter=<service>` for targeted builds
- **Nx**: Use `nx build <service>` or `nx affected --target=build` for selective builds
- Both tools handle dependency ordering automatically
- In Docker: install the build orchestrator during the build stage, use it for targeted builds

## Gotchas

- pnpm hoists differently than npm/yarn — `--shamefully-hoist` may be needed for some packages
- Root `tsconfig.json` with `references` must be present for TypeScript project references to work
- Turborepo `--filter` uses package names from `package.json`, not directory names
- Go workspaces: `go.work.sum` must also be committed alongside `go.work`
- Nx `affected` needs git history in the Docker build — use `--base=HEAD~1` or copy `.git` (adds size)
- Each service in a compose file can have different environment variables — don't share `.env` across services that need different configs

## Related Skills

- **`compose-setup`** — Base compose patterns for databases and caches
- **`dockerfile-generation`** — Ecosystem-specific Dockerfile patterns to adapt for monorepo services
- **`node-deploy`** — Node.js monorepo support (workspaces, package managers)
- **`go-deploy`** — Go workspace support

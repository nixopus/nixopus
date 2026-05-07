---
name: deno-deploy
description: Build and deploy Deno applications — version detection, dependency caching, and Dockerfile patterns. Use when deploying a Deno project, or when deno.json or deno.jsonc is detected.
metadata:
  version: "1.0"
---

# Deno Deployment

## Detection

Project is Deno if `deno.json` or `deno.jsonc` exists in the root.

## Versions

1. `.deno-version` file or `mise.toml` / `.tool-versions`
2. Defaults to **2**

## Build

### Build Process

1. Install Deno
2. Cache dependencies: `deno cache`
3. Start command derived from project config

### Start Command

1. `main.ts`, `main.js`, `main.mjs`, or `main.mts` in project root
2. If none found, use first `.ts`, `.js`, `.mjs`, or `.mts` file
3. Run with: `deno run --allow-all <entry>`

`deno.json` / `deno.jsonc` may specify `tasks` or `main`; prefer those when present.

## Install Stage Optimization

Copy in order:
- `deno.json`, `deno.jsonc`
- `lock.json` (if present)
- `*.ts`, `*.js`, `*.mjs`, `*.mts` (or full source)

## Dockerfile Patterns

### Basic Deno App

```dockerfile
FROM denoland/deno:2
WORKDIR /app
COPY deno.json deno.jsonc ./
COPY . .
RUN deno cache main.ts
EXPOSE 3000
CMD ["deno", "run", "--allow-all", "main.ts"]
```

### With lock file

```dockerfile
FROM denoland/deno:2
WORKDIR /app
COPY deno.json deno.jsonc lock.json ./
COPY src/ ./src/
RUN deno cache src/main.ts
COPY . .
EXPOSE 3000
CMD ["deno", "run", "--allow-all", "src/main.ts"]
```

### deno.json tasks

```dockerfile
FROM denoland/deno:2
WORKDIR /app
COPY deno.json deno.jsonc ./
COPY . .
RUN deno cache main.ts
EXPOSE 3000
CMD ["deno", "task", "start"]
```

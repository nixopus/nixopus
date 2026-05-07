---
name: gleam-deploy
description: Build and deploy Gleam applications — erlang-shipment, version detection, and Dockerfile patterns. Use when deploying a Gleam project, or when gleam.toml is detected.
metadata:
  version: "1.0"
---

# Gleam Deployment

## Detection

Project is Gleam if `gleam.toml` exists in the root.

## Versions

Gleam and Erlang default to latest. Override via `mise.toml` or `.tool-versions`.

Erlang available in both build and runtime; Gleam only during build.

## Build

### Build Process

1. Install Gleam and Erlang
2. Export: `gleam export erlang-shipment`
3. Output: `./build/erlang-shipment/`

### Start Command

```
./build/erlang-shipment/entrypoint.sh run
```

Source tree not included in final container by default.

## Install Stage Optimization

Copy in order:
- `gleam.toml`, `manifest.toml` (if present)
- `src/` (Gleam source)
- `test/` (optional)

## Base Images

| Stage | Image |
|---|---|
| Build | `ghcr.io/gleam-lang/gleam:latest` or custom |
| Runtime | `erlang:27-slim` |

## Dockerfile Patterns

### Erlang Shipment (minimal runtime)

```dockerfile
FROM ghcr.io/gleam-lang/gleam:latest AS build
WORKDIR /app
COPY gleam.toml manifest.toml* ./
COPY src src
RUN gleam export erlang-shipment

FROM erlang:27-slim
WORKDIR /app
COPY --from=build /app/build/erlang-shipment ./
EXPOSE 8080
CMD ["./entrypoint.sh", "run"]
```

### With source included

```dockerfile
FROM ghcr.io/gleam-lang/gleam:latest AS build
WORKDIR /app
COPY . .
RUN gleam export erlang-shipment

FROM erlang:27-slim
WORKDIR /app
COPY --from=build /app/build/erlang-shipment ./
COPY --from=build /app/src ./src
COPY --from=build /app/gleam.toml ./
CMD ["./entrypoint.sh", "run"]
```

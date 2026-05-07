---
name: rust-deploy
description: Build and deploy Rust applications — version detection, release binaries, cargo-chef, and Dockerfile patterns. Use when deploying a Rust project, or when Cargo.toml is detected.
metadata:
  version: "1.0"
---

# Rust Deployment

## Detection

Project is Rust if `Cargo.toml` or `cargo.toml` exists.

## Versions

Rust version priority:

1. `rust-toolchain.toml` → `toolchain.channel`
2. `Cargo.toml` → `package.rust-version`
3. `.rust-version` file
4. `Cargo.toml` → `package.edition` (maps to minimum Rust version)
5. Defaults to latest stable (**1.89**)

## Build

### Build Command

- `cargo build --release`
- Binary: `target/release/<package-name>`
- Package name from `Cargo.toml` → `[package].name`

### Entry Resolution

1. Binary crate: `src/main.rs`
2. Multi-binary: `src/bin/<name>.rs` → binary name matches filename
3. Library + binary: `[bin]` section in `Cargo.toml`

### Output

- Build output: `target/release/<name>`
- Start command: `./bin/<project-name>` or `./target/release/<name>`

## Runtime Variables

| Variable | Purpose |
|---|---|
| `ROCKET_ADDRESS` | Rocket bind address (use `0.0.0.0` for containers) |
| `PORT` | Port for web servers (default 8080) |

## Port Detection

1. `PORT` in `.env` / `.env.example`
2. Source code patterns: `:8080`, `bind`, `TcpListener` in `main.rs`
3. Default: **8080**

## Framework / Library Detection

| Crate | Category |
|---|---|
| `axum` | Web |
| `actix-web` | Web |
| `rocket` | Web |
| `warp` | Web |
| `tokio` | Async runtime |
| `tower` | Middleware |

## Install Stage Optimization

Copy in order:
- `Cargo.toml`, `Cargo.lock`
- `src/` (or full source)

## BuildKit Caching

| Cache key | Path |
|---|---|
| `cargo_registry` | `~/.cargo/registry` |
| `cargo_git` | `~/.cargo/git` |
| `cargo_target` | `target/` |

## Base Images

| Stage | Image |
|---|---|
| Build | `rust:1.89-bookworm` or `rust:1.89-alpine` |
| Runtime | `debian:bookworm-slim` or `gcr.io/distroless/cc` |

For Alpine build: may need `musl-dev` for fully static binary.

## Dockerfile Patterns

### Multi-Stage (Debian)

```dockerfile
FROM rust:1.89-bookworm AS build
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
COPY src ./src
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/target/release/<binary-name> .
EXPOSE 8080
CMD ["./<binary-name>"]
```

### Cargo Chef (faster layer cache)

```dockerfile
FROM rust:1.89-bookworm AS chef
RUN cargo install cargo-chef

FROM chef AS planner
WORKDIR /app
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

FROM chef AS build
WORKDIR /app
COPY --from=planner /app/recipe.json .
RUN cargo chef cook --release --recipe-path recipe.json
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/target/release/<binary-name> .
EXPOSE 8080
CMD ["./<binary-name>"]
```

### Simple (no chef)

```dockerfile
FROM rust:1.89-bookworm AS build
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/target/release/<binary-name> .
EXPOSE 8080
CMD ["./<binary-name>"]
```

## Gotchas

- The dummy `main.rs` trick (`echo "fn main() {}" > src/main.rs`) caches dependency compilation, but `touch src/main.rs` is needed afterward to invalidate the binary build cache
- `Cargo.lock` must be committed for binary crates — without it, dependency versions are resolved fresh on every build
- Alpine builds need `musl-dev`, but crates using OpenSSL also need `openssl-dev` and `pkgconfig`
- Rust builds are memory-intensive — `cargo build --release -j 2` limits parallelism to avoid OOM kills in resource-constrained environments
- Cargo chef `prepare` requires the full source tree — it scans code to compute the dependency recipe

---
name: go-deploy
description: Build and deploy Go applications — version detection, static binaries, CGO, workspaces, and Dockerfile patterns. Use when deploying a Go project, or when go.mod is detected.
metadata:
  version: "1.0"
---

# Go Deployment

## Detection

Project is Go if any of these exist:

- `go.mod` at the build context root
- `go.work` at the root (Go workspaces)
- `main.go` at the root

## Versions

Go version priority:

1. `go.mod` → `go` directive (e.g. `go 1.21`)
2. `.go-version` file, `mise.toml`, or `.tool-versions`
3. Defaults to **1.23**

## Build

### Build Command

- Single module: `go build -ldflags="-w -s" -o /app/out .`
- CGO disabled (static): `CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/out .`
- To build a specific binary from `cmd/`, target it directly: `go build -ldflags="-w -s" -o /app/<name> ./cmd/<name>`
- For workspaces, specify the module path: `go build -ldflags="-w -s" -o /app/out ./<module>`

### Main Package Resolution

1. Root directory if it contains `.go` files
2. First subdirectory in `cmd/` (e.g. `./cmd/server`)
3. For workspaces (`go.work`): first module with a `main` package

### Output

- Binary: `/app/out` or `/app/<name>`
- Static binary when `CGO_ENABLED=0`

## Go Workspaces

For multi-module projects with `go.work`:

- Discovers and copies all module dependencies
- Builds first module with `main.go` by default
- Specify the target module path to build a different one

## CGO Support

Default: CGO disabled (`CGO_ENABLED=0`) for static binaries. If CGO needed:

- Set `CGO_ENABLED=1`
- Build stage needs: `gcc`, `g++`, `libc6-dev`
- Runtime needs `libc6` for dynamic linking
- Use `debian:bookworm-slim` instead of `alpine` for runtime

## Port Detection

1. `PORT` in `.env` / `.env.example`
2. Source code patterns: `:8080`, `ListenAndServe`, `Run(` in `main.go`
3. Default: **8080**

## Framework / Library Detection

| Import / package | Category |
|---|---|
| `net/http` | Standard library |
| `github.com/gin-gonic/gin` | Gin |
| `github.com/labstack/echo` | Echo |
| `github.com/go-chi/chi` | Chi |
| `github.com/valyala/fasthttp` | FastHTTP |
| `github.com/gofiber/fiber` | Fiber |

## Install Stage Optimization

Copy in order:
- `go.mod`, `go.sum` (and `go.work` if workspace)
- Workspace: copy all module directories referenced in `go.work`
- `*.go` (or full source for multi-package)

Copy `go.mod` + `go.sum` first for layer caching.

## Caching

Use BuildKit cache mount:

```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" -o /app/out .
```

## Base Images

| Stage | Image |
|---|---|
| Build | `golang:1.23-alpine` or `golang:1.23-bookworm` |
| Runtime (static) | `gcr.io/distroless/static` or `alpine:latest` |
| Runtime (CGO) | `debian:bookworm-slim` |

## Dockerfile Patterns

### Static Binary (Alpine runtime)

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/out .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/out .
EXPOSE 8080
CMD ["./out"]
```

### Distroless Runtime

```dockerfile
FROM golang:1.23-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/out .

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /app/out .
EXPOSE 8080
CMD ["./out"]
```

### cmd layout

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Go workspace

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.work go.work.sum ./
COPY api/go.mod api/go.sum ./api/
COPY shared/go.mod shared/go.sum ./shared/
RUN go mod download ./api/...
COPY api/ ./api/
COPY shared/ ./shared/
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/out ./api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/out .
EXPOSE 8080
CMD ["./out"]
```

## Gotchas

- `CGO_ENABLED=0` is required for truly static binaries — without it, the binary may dynamically link glibc and fail on Alpine/distroless
- `go.sum` must be committed to the repo — missing it causes `go mod download` to fail in Docker
- Multi-binary repos with `cmd/` layout require specifying the target: `go build ./cmd/server`, not just `go build .`
- Alpine runtime images need `ca-certificates` for outbound HTTPS — distroless/static includes them by default
- Go modules cache at `/go/pkg/mod` — use BuildKit cache mounts to avoid re-downloading on every build

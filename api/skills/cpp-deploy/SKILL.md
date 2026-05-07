---
name: cpp-deploy
description: Build and deploy C/C++ applications — CMake, Meson, Ninja, and Dockerfile patterns. Use when deploying a C or C++ project, or when CMakeLists.txt or meson.build is detected.
metadata:
  version: "1.0"
---

# C/C++ Deployment

## Detection

Project is C/C++ if `CMakeLists.txt` or `meson.build` exists in the root.

## Versions

Latest CMake (or Meson) and Ninja installed during build. No explicit version pinning by default.

## Build

### Build Process

1. Install CMake or Meson and Ninja
2. Configure and build into `./build` (or `/build`)
3. Executable placed in the build directory

### Start Command

Executable name matches the project root directory name, located in the build directory.
Example: project `my_app/` → `./build/my_app` or `/build/my_app`.

### Output

- Build directory: `/build` (or `./build`)
- Source tree not included in final container; only build artifacts

## Build Systems

| File | Build system |
|---|---|
| `CMakeLists.txt` | CMake |
| `meson.build` | Meson |

Both use Ninja as the default generator/backend.

## Install Stage Optimization

Copy in order:
- `CMakeLists.txt` or `meson.build`
- `meson.options` (Meson)
- `src/`, `include/`, or full source tree

## Base Images

| Stage | Image |
|---|---|
| Build | `gcc` or `clang` + CMake/Meson + Ninja |
| Runtime | `debian:bookworm-slim` or `alpine` (copy binary only) |

## Dockerfile Patterns

### CMake

```dockerfile
FROM gcc:13-bookworm AS build
RUN apt-get update && apt-get install -y cmake ninja-build && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY CMakeLists.txt ./
COPY src src
RUN cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release && cmake --build build

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates libstdc++6 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/build/my_app .
EXPOSE 8080
CMD ["./my_app"]
```

### Meson

```dockerfile
FROM gcc:13-bookworm AS build
RUN apt-get update && apt-get install -y meson ninja-build && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY meson.build meson.options* ./
COPY src src
RUN meson setup build -Dbuildtype=release && ninja -C build

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates libstdc++6 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/build/my_app .
EXPOSE 8080
CMD ["./my_app"]
```

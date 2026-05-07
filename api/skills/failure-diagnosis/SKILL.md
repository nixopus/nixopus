---
name: failure-diagnosis
description: Diagnose deployment failures, container crashes, and networking issues using structured pattern matching on logs and container state. Use when a deployment fails, a container crashes or exits unexpectedly, or the app is unreachable after deployment.
metadata:
  version: "1.0"
---

# Failure Diagnosis

When diagnosing a deployment issue, work through the relevant section based on the symptom. Use the pattern tables to match log output before hypothesizing.

## Build Failure Patterns

After calling `get_deployment_logs`, scan the output for these patterns.

### Node.js

| Log pattern | Root cause | Fix |
|---|---|---|
| `ENOMEM` or `JavaScript heap out of memory` | Node ran out of memory during build | Add `NODE_OPTIONS=--max-old-space-size=4096` as build-time env var |
| `ERR_MODULE_NOT_FOUND` or `Cannot find module` | Missing dependency or wrong import path | Check `package.json` dependencies; verify the module is not dev-only if pruned |
| `error TS` followed by file path and line number | TypeScript compilation error | Read the referenced file — usually a type mismatch or missing type package |
| `sharp: Installation error` or `something went wrong installing the "sharp" package` | Missing `libvips` system dependency | Add `RUN apk add --no-cache vips-dev` (alpine) or `RUN apt-get install -y libvips-dev` (debian) before `npm install` |
| `gyp ERR!` or `node-gyp rebuild` | Native addon compilation failed — missing `python3`, `make`, or `g++` | Add build tools: `RUN apk add --no-cache python3 make g++` (alpine) or `RUN apt-get install -y python3 make g++` (debian) |
| `npm warn ERESOLVE` or `Could not resolve dependency` | Dependency version conflict | Add `--legacy-peer-deps` to install command, or fix the conflicting version ranges |
| `Error: EACCES: permission denied` | Dockerfile runs as non-root but writes to root-owned directory | Add `RUN chown -R node:node /app` before switching to non-root user |
| `next build` fails with `Module not found` for `@/` paths | Next.js path alias not resolving | Verify `tsconfig.json` `paths` and that all source files are copied before build |
| `.next/standalone` directory missing after build | `output: "standalone"` not set in `next.config.*` | Add `output: "standalone"` to Next.js config |

### Python

| Log pattern | Root cause | Fix |
|---|---|---|
| `ModuleNotFoundError: No module named` | Package not in requirements or venv not activated | Verify the module is listed in `requirements.txt` / `pyproject.toml`; check Dockerfile uses the same Python that pip installed to |
| `error: subprocess-exited-with-error` during pip install | Native extension compilation failed | Install system build deps: `RUN apt-get install -y build-essential libpq-dev libffi-dev` |
| `pg_config executable not found` | `psycopg2` needs PostgreSQL client libs | Use `psycopg2-binary` instead, or install `libpq-dev` |
| `Could not find a version that satisfies` | Pip version conflict or typo in package name | Check package name spelling and version constraints |
| `RuntimeError: uvloop does not support Windows` | Wrong base image platform | Ensure Dockerfile uses a linux base image |
| `Permission denied: '/app'` | Non-root user can't write to workdir | Add `RUN chown -R appuser:appuser /app` |

### Go

| Log pattern | Root cause | Fix |
|---|---|---|
| `cannot find module providing package` | Missing dependency or wrong module path | Run `go mod tidy` — the go.sum may be stale |
| `cgo: C compiler "gcc" not found` | CGO enabled but no C compiler in image | Either `CGO_ENABLED=0` for static builds, or install `gcc` and `musl-dev` |
| `signal: killed` during build | OOM during compilation | Increase build memory or reduce parallelism with `-p 1` flag |
| `main.go:X: undefined:` | Function or variable not exported or wrong package | Check capitalization (Go exports are uppercase) and build tags |

### Rust

| Log pattern | Root cause | Fix |
|---|---|---|
| `error[E0433]: failed to resolve` | Missing crate or wrong import path | Check `Cargo.toml` dependencies |
| `Killed` or `signal: 9` during `cargo build` | OOM during compilation — Rust builds are memory-intensive | Use `cargo build --release -j 2` to limit parallelism, or increase build memory |
| `linking with cc failed` | Missing system libraries for C bindings | Install required `-dev` packages (e.g., `libssl-dev`, `pkg-config`) |
| `error: linker cc not found` | No C linker in image | Install `build-essential` or `gcc` |

### Java

| Log pattern | Root cause | Fix |
|---|---|---|
| `java.lang.OutOfMemoryError: Java heap space` | Maven/Gradle OOM during build | Set `MAVEN_OPTS=-Xmx1024m` or `GRADLE_OPTS=-Xmx1024m` |
| `ERROR: JAVA_HOME is not set` | JDK not installed or JAVA_HOME not configured | Ensure Dockerfile uses a JDK base image (not JRE) for build stage |
| `Could not find artifact` | Missing Maven dependency or wrong repository URL | Check `pom.xml` repositories and dependency coordinates |
| `Compilation failure` with source/target version | Java source version mismatch | Match `source`/`target` in pom.xml to the JDK version in the base image |

### General Dockerfile

| Log pattern | Root cause | Fix |
|---|---|---|
| `COPY failed: file not found in build context` | Source path in `COPY` doesn't exist, or `.dockerignore` excludes it | Verify the path exists and is not in `.dockerignore` |
| `failed to solve: not found` after `FROM ... AS` | Multi-stage stage name typo or missing stage | Check that the stage name in `COPY --from=` matches a `FROM ... AS` stage |
| `manifest unknown` or `not found` in registry | Base image tag doesn't exist | Verify the image:tag exists on Docker Hub / registry |
| `executor failed running`: `No such file or directory` | Script referenced in `CMD`/`ENTRYPOINT` doesn't exist or has wrong line endings | Check the file exists in the final stage; convert CRLF to LF if built on Windows |
| `Error response from daemon: conflict` | Container name already in use | Previous deployment didn't clean up — remove the old container first |

## Container Runtime Failures

After a deployment succeeds (image built) but the container crashes or misbehaves.

### Exit Codes

| Exit code | Signal | Meaning |
|---|---|---|
| 0 | — | Clean exit — process finished normally (unexpected for a long-running server) |
| 1 | — | Generic application error — check application logs |
| 126 | — | Command found but not executable — check file permissions on entrypoint |
| 127 | — | Command not found — entrypoint binary missing from final image stage |
| 137 | SIGKILL (9) | Killed externally — usually OOM killer or `docker stop` timeout |
| 139 | SIGSEGV (11) | Segmentation fault — native code crash, corrupted memory |
| 143 | SIGTERM (15) | Graceful termination — normal `docker stop` |

Exit codes 128+N mean the process was killed by signal N.

### Container Inspect Signals

Use `container_inspect` to check these fields:

| Field | Condition | Meaning |
|---|---|---|
| `oom_killed` | `true` | Container exceeded memory limit — increase memory or fix memory leak |
| `restart_count` | > 5 in last hour | Crash loop — container starts, crashes, restarts repeatedly |
| `health_status` | `unhealthy` | Healthcheck endpoint failing — check if the app's health endpoint is reachable inside the container |
| `health_status` | `starting` for > 60s | App takes too long to boot — slow startup or stuck initialization |

### Common Runtime Patterns

Scan `get_container_logs` or `get_application_logs` for these:

| Log pattern | Root cause | Fix |
|---|---|---|
| `EADDRINUSE` or `address already in use` | Port conflict — another process holds the port | Check for duplicate containers, or the app spawns a child that binds first |
| `ECONNREFUSED` to database host | Database not reachable from container network | Verify DB host is correct for Docker networking (use service name, not `localhost`) |
| `undefined` or `TypeError: Cannot read properties of undefined` (Node) | Missing environment variable | Cross-reference app env var usage with configured vars via `container_exec ["env"]` |
| `KeyError` or `os.environ` error (Python) | Missing environment variable | Same — check configured env vars |
| `ENOENT: no such file or directory` | Expected file not in container | Verify `COPY` in Dockerfile includes the file; check `.dockerignore` |
| `permission denied` on file operations | Non-root user lacks write access | `chown` the directory in Dockerfile or use a writable volume |
| `FATAL: password authentication failed` | Wrong database credentials | Verify `DATABASE_URL` or individual DB credential env vars |
| `EMFILE: too many open files` | File descriptor limit reached | Add `ulimits` in compose or increase container fd limit |
| `ERR_DLOPEN_FAILED` (Node) | Native module compiled for wrong architecture | Rebuild native modules inside the Docker build (don't copy host `node_modules`) |

### Crash Loop Detection

A container is in a crash loop when:
1. `restart_count` > 3 in the last 10 minutes
2. Container logs show the same error repeating
3. Container uptime resets to 0 repeatedly

To diagnose a crash loop:
1. `get_container_logs` for the last 100 lines
2. `container_inspect` for `oom_killed`, exit code, restart count
3. If OOM: `container_stats` to see memory usage trend
4. If exit 1: search logs for the first error after startup
5. If exit 137 but not OOM: check host memory via machine-agent delegation

## Networking Failures

When the app runs but is not reachable externally.

### Port Mismatch Diagnosis

Four values must agree — a mismatch at any level causes unreachable apps:

| Layer | How to check | Tool |
|---|---|---|
| App listen port | `container_exec ["ss", "-tlnp"]` or grep source for `.listen(` | `container_exec` |
| Dockerfile EXPOSE | `container_inspect` → `ports` | `container_inspect` |
| Nixopus app config port | `get_application` → port field | `get_application` |
| Proxy upstream port | `proxy_config` → `upstream` | `proxy_config` |

If any disagree, the app is unreachable. The app listen port is the source of truth — all others must match it.

### Reachability Matrix

Use this decision tree when the app is reported as unreachable:

| Check | Tool | Pass means | Fail means |
|---|---|---|---|
| External URL responds | `http_probe` on public URL | App is reachable (problem may be intermittent) | Continue to next check |
| App responds inside container | `container_exec ["curl", "-s", "localhost:PORT"]` | App is running; problem is proxy/DNS/network | App itself is down — check container logs |
| Container is running | `list_containers` / `get_container` | Container exists and is up | Container crashed — see Container Runtime Failures |
| DNS resolves to server | `network_diagnostics` with type `dns` | Domain points to correct IP | DNS misconfigured — check domain settings |
| Port is open on host | `network_diagnostics` with type `port` | Traffic reaches the server | Firewall or port not published |

### Proxy and TLS Issues

| Symptom | Root cause | Diagnosis |
|---|---|---|
| 502 Bad Gateway | Proxy can't reach upstream container | `proxy_config` to check upstream; `container_exec` curl to verify app is listening |
| 503 Service Unavailable | App overloaded or not ready | `container_stats` for CPU/memory; check if app has finished starting |
| 504 Gateway Timeout | Upstream too slow to respond | App may be stuck — `container_exec ["ps", "aux"]` to check for hung processes |
| `SSL_ERROR` or `ERR_CERT_AUTHORITY_INVALID` | TLS certificate issue | `proxy_config` to check `tls_enabled`; Caddy auto-TLS may need valid DNS first |
| Redirect loop (ERR_TOO_MANY_REDIRECTS) | App and proxy both forcing HTTPS redirect | Disable app-level HTTPS redirect — let the proxy handle TLS termination |

### Container-to-Service Connectivity

When the app can't reach its dependencies (database, cache, external API):

1. `container_exec ["nslookup", "<hostname>"]` — DNS resolution
2. `container_exec ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "<url>"]` — HTTP reachability
3. `network_diagnostics` with type `port` — TCP connectivity
4. Check if both containers are on the same Docker network via `container_inspect` → `networks`

## Diagnostic Decision Tree

Start here. Match the symptom, follow the path.

### Symptom: `build_failed`

1. `get_application_deployments` to find the failed deployment
2. `get_deployment_logs` for the full build output
3. Scan logs against **Build Failure Patterns** tables above
4. If match found: apply the documented fix
5. If no match: search for the first `error` or `Error` line — that's usually the root cause (earlier lines are often cascading failures)

### Symptom: container exited / crash loop

1. `get_application` to confirm deployment status
2. `list_containers` to find the container (it may have been removed on crash)
3. `container_inspect` for exit code, `oom_killed`, `restart_count`, `health_status`
4. Map exit code using **Exit Codes** table
5. `get_container_logs` (or `get_application_logs` if container is gone) for the last error
6. Match log output against **Common Runtime Patterns** table
7. If OOM: `container_stats` to see current memory usage vs limit

### Symptom: app unreachable

1. `http_probe` the public URL — if it responds, problem is intermittent or resolved
2. `get_application` to confirm the app exists and has a deployment
3. `list_containers` to check container is running
4. `container_exec ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "localhost:PORT"]` — internal reachability
5. If internal works but external doesn't: `proxy_config` for upstream mismatch, then **Port Mismatch Diagnosis**
6. If internal fails too: check container logs for startup errors
7. If container is running but not listening: `container_exec ["ss", "-tlnp"]` to see what ports are bound

### Symptom: intermittent errors / slow responses

1. `container_stats` for CPU and memory pressure
2. `get_container_logs` with recent timeframe — look for error spikes
3. If memory > 80% of limit: approaching OOM, recommend increasing memory or fixing leak
4. If CPU > 90%: app is compute-bound, check for infinite loops or missing caching
5. `http_probe` multiple times to confirm intermittent pattern
6. Check `container_inspect` → `restart_count` for silent crash-restarts

## Related Skills

- **`domain-tls-routing`** — For domain resolution, TLS certificate, and reverse proxy routing issues specifically
- **`pre-deploy-checklist`** — Run before deployment to catch issues that would cause the failures diagnosed here

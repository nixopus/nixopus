---
name: container-resource-tuning
description: Size container memory and CPU limits, diagnose OOM kills and CPU throttling, and recommend resource adjustments by ecosystem. Use when containers are being OOM-killed, running slowly, or when setting initial resource limits for a deployment.
metadata:
  version: "1.0"
---

# Container Resource Tuning

## Default Resource Recommendations

Starting points by ecosystem. Adjust based on actual usage.

| Ecosystem | Memory limit | CPU shares | Notes |
|---|---|---|---|
| Node.js | 512MB | 0.5 | V8 GC is memory-hungry; Next.js SSR needs more |
| Node.js (Next.js SSR) | 1024MB | 1.0 | Server-side rendering is CPU and memory intensive |
| Python (Django/Flask) | 512MB | 0.5 | Per-worker; multiply by worker count |
| Python (FastAPI) | 256MB | 0.5 | Async, lower per-process memory |
| Go | 256MB | 0.5 | Static binary, efficient memory use |
| Rust | 128MB | 0.25 | Minimal runtime overhead |
| Java (Spring Boot) | 1024MB | 1.0 | JVM needs headroom; set `-Xmx` to 75% of limit |
| PHP (FrankenPHP) | 512MB | 0.5 | Per-request memory; depends on payload |
| Ruby (Rails) | 512MB | 0.5 | Per-worker; Puma workers multiply this |
| Elixir (Phoenix) | 256MB | 0.5 | BEAM VM is efficient; handles concurrency well |
| .NET (ASP.NET) | 512MB | 0.5 | Similar to Node.js profile |
| Static (Caddy/nginx) | 64MB | 0.25 | Minimal; just serving files |

## Diagnosing OOM Kills

When `container_inspect` shows `oom_killed: true`:

1. **Check current limit**: `container_inspect` → memory limit
2. **Check peak usage**: `container_stats` → memory usage and limit
3. **Check what's consuming memory**:
   - `container_exec ["ps", "aux", "--sort=-%mem"]` → top processes
   - Node.js: `container_exec ["node", "-e", "console.log(process.memoryUsage())"]`

### Common causes

| Ecosystem | Cause | Fix |
|---|---|---|
| Node.js | V8 heap exceeds limit | Set `NODE_OPTIONS=--max-old-space-size=<MB>` to 75% of container limit |
| Node.js | Memory leak (heap grows unbounded) | Profile with `--inspect`; check for event listener leaks, unbounded caches |
| Java | JVM default heap exceeds container limit | Set `-Xmx` to 75% of container memory limit |
| Python | Large dataset loaded into memory | Use streaming/chunked processing; increase limit if data size is fixed |
| Any | Too many worker processes | Reduce worker count: Gunicorn `--workers`, Puma `workers`, PM2 instances |

### Right-sizing after OOM

1. Increase memory limit by 50% from current value
2. Deploy and monitor `container_stats` for 10 minutes
3. If peak usage is consistently below 60% of limit: limit is right
4. If peak usage exceeds 80%: increase again or investigate the memory consumer
5. If peak usage is below 30%: reduce limit to save resources

## Diagnosing CPU Throttling

When the app is slow but not OOM-killed:

1. **Check CPU usage**: `container_stats` → CPU percentage
2. **Check host load**: `get_machine_stats` → system load average
3. **Check for CPU-bound work**:
   - `container_exec ["ps", "aux", "--sort=-%cpu"]` → top CPU consumers

### Common causes

| Symptom | Cause | Fix |
|---|---|---|
| CPU at 100% of limit | App is compute-bound | Increase CPU shares or optimize hot paths |
| CPU at 100%, response times spike | Not enough CPU for request volume | Scale horizontally (more instances) or increase CPU |
| Low CPU but slow responses | Waiting on I/O (database, external API) | Not a CPU issue — check database latency |
| Host load > 2x cores | Server overloaded | Multiple containers competing — reduce total load or upgrade server |

## JVM-Specific Tuning

Java apps need explicit JVM flags to respect container limits:

```
JAVA_TOOL_OPTIONS=-XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0
```

- `UseContainerSupport` (default since Java 10): JVM reads cgroup memory limits
- `MaxRAMPercentage=75.0`: heap uses 75% of container memory, leaving room for native memory and GC

## Node.js-Specific Tuning

```
NODE_OPTIONS=--max-old-space-size=384
```

For a 512MB container, set old space to ~75% (384MB). V8 needs headroom for GC, native code, and buffers.

For production, also set:
- `UV_THREADPOOL_SIZE=4` (default) — increase for I/O-heavy apps
- `NODE_CLUSTER_WORKERS` — if using cluster mode, each worker needs its own memory budget

## Python-Specific Tuning

Gunicorn workers multiply memory usage:

```
gunicorn app:app --workers 2 --worker-class uvicorn.workers.UvicornWorker
```

Rule of thumb: `workers = (2 * CPU cores) + 1`, but in containers with limited CPU, use 2-4 workers max.

Each worker uses roughly the same memory as a single process. 4 workers × 256MB = 1GB total.

## Compose Resource Limits

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

- `limits`: hard ceiling — container is OOM-killed if exceeded
- `reservations`: guaranteed minimum — Docker ensures this is available

## Monitoring After Changes

After adjusting resources:

1. `container_stats` — check memory and CPU usage over time
2. `get_container_logs` — scan for OOM warnings or performance errors
3. `http_probe` — verify response times are acceptable
4. If `restart_count` drops to 0 and memory stays below 80%: tuning is correct

## Related Skills

- **`post-deploy-verification`** — Check container stability after resource changes
- **`failure-diagnosis`** — Exit code 137 (OOM kill) diagnosis
- **`compose-setup`** — Resource limits in docker-compose.yml

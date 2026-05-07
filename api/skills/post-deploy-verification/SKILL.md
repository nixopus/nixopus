---
name: post-deploy-verification
description: Verify a deployment is healthy after it completes — HTTP probes, healthcheck endpoints, container stability, log scanning, and port alignment. Use after any deployment to confirm the app is running and reachable.
metadata:
  version: "1.0"
---

# Post-Deploy Verification

Run these checks in order after a deployment completes. Report all results — do not stop at the first failure.

## Step 1: Container is running

- `list_containers` to find the app's container
- Verify status is `running` (not `exited`, `restarting`, `created`)
- If container is missing or exited, skip remaining steps — the deployment failed at container level

## Step 2: No restart loop

- `container_inspect` → check `restart_count`
- If `restart_count` > 0 within 60 seconds of deployment: container is crash-looping
- Check `oom_killed` — if true, the container exceeded its memory limit

## Step 3: Port alignment

Four values must agree:

| Layer | How to check |
|---|---|
| App listen port | `container_exec ["ss", "-tlnp"]` or grep source for `.listen(` |
| Dockerfile EXPOSE | `container_inspect` → `ports` |
| App config port | `get_application` → port field |
| Proxy upstream | `proxy_config` → `upstream` |

If any disagree, the app will be unreachable even though the container is running.

## Step 4: Internal reachability

- `container_exec ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "localhost:PORT"]`
- Expect 200, 301, or 302
- If connection refused: app hasn't started listening yet (may need to wait) or wrong port
- If timeout: app is hanging during startup

## Step 5: External reachability

- `http_probe` the public URL
- Expect 200 (or 301/302 for SPAs with redirect)
- If internal works but external fails: proxy/DNS/TLS issue — defer to domain-tls-routing

## Step 6: Healthcheck endpoint

If the app has a healthcheck endpoint (`/health`, `/healthz`, `/api/health`, `/ready`):

- `container_exec ["curl", "-s", "localhost:PORT/health"]`
- Parse response: look for `"status": "ok"` or `"healthy"` or HTTP 200
- If unhealthy: the app started but a dependency (database, cache, external service) is down

## Step 7: Log scan

- `get_container_logs` — last 50 lines
- Scan for error patterns:

| Pattern | Meaning |
|---|---|
| `ECONNREFUSED` | Database or service not reachable |
| `EADDRINUSE` | Port conflict |
| `Error:` or `FATAL` | Application error during startup |
| `TypeError` / `ReferenceError` (Node) | Code error |
| `ModuleNotFoundError` (Python) | Missing dependency |
| `panic:` (Go) | Runtime panic |

- No errors in first 50 lines after startup = healthy

## Step 8: Compose services (if applicable)

For docker-compose deployments:

- `get_compose_services` to list all services
- Run steps 1-7 for each service independently
- Verify service-to-service connectivity: primary app can reach its database/cache

## Result format

| Check | Status | Details |
|-------|--------|---------|
| Container running | PASS/FAIL | Container ID and status |
| No restart loop | PASS/FAIL | restart_count, oom_killed |
| Port alignment | PASS/FAIL | Expected vs actual |
| Internal reachable | PASS/FAIL | HTTP status code |
| External reachable | PASS/FAIL | HTTP status code |
| Healthcheck | PASS/WARN/N/A | Endpoint and response |
| Log scan | PASS/FAIL | Error patterns found |
| Compose services | PASS/FAIL/N/A | Service health summary |

**Healthy**: All checks PASS (or WARN/N/A for optional checks).
**Unhealthy**: Any FAIL — report the first failing check as the likely root cause.

## Related Skills

- **`failure-diagnosis`** — If verification fails, use failure diagnosis for deeper investigation
- **`domain-tls-routing`** — If internal reachability passes but external fails

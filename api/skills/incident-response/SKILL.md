---
name: incident-response
description: Structured incident response workflow — severity classification, diagnosis delegation, auto-fix decisions, notification, and post-incident review. Use when an automated failure event is received or when the user reports a production incident.
metadata:
  version: "1.0"
---

# Incident Response

## Event Classification

Classify the incident by severity before acting:

| Severity | Criteria | Response time | Action |
|---|---|---|---|
| **Critical** | App completely down, all users affected | Immediate | Diagnose + attempt auto-fix + notify |
| **High** | App degraded, errors for some users | Within minutes | Diagnose + attempt auto-fix + notify |
| **Medium** | Non-user-facing failure (build failed, deploy failed) | Within session | Diagnose + fix suggestion + notify |
| **Low** | Warning, non-critical issue detected | Informational | Notify only |

### Severity signals

| Signal | Severity |
|---|---|
| Container exited, `restart_count` > 3 | Critical |
| HTTP probe returns 502/503/504 | Critical |
| Build failed | Medium |
| Container OOM-killed once | High |
| Health endpoint returns unhealthy | High |
| Deployment succeeded but no traffic | High |
| SSL certificate expiring soon | Medium |
| Disk usage > 85% | Medium |

## Incident Workflow

### 1. GATHER

Collect context about the affected resource:

- `get_application` — app details, current deployment, configured port
- `get_application_deployments` — recent deployment history
- `get_deployment_logs` — if build/deploy failed
- `list_containers` → `get_container` — container status
- `get_container_logs` — runtime errors

### 2. DIAGNOSE

Delegate to the diagnostic agent with full context:

- Include: application ID, deployment ID, error message, container status
- The diagnostic agent uses `failure-diagnosis` skill for pattern matching
- Wait for diagnosis result: root cause + whether it's code-fixable

### 3. DECIDE

Based on diagnosis:

| Root cause type | Action |
|---|---|
| Code error (syntax, missing dep, config) | Auto-fix via PR |
| Dockerfile issue (wrong base image, missing file) | Auto-fix via PR |
| Environment variable missing or wrong | Notify user — env vars need manual input |
| Infrastructure (server resources, Docker daemon) | Notify user — requires manual intervention |
| Database connection failed | Notify user — check database status and credentials |
| External service down | Notify user — nothing to fix on our side |
| Unknown | Notify user with gathered evidence |

### 4. FIX (if code-fixable)

Delegate to the GitHub agent:

- Branch: `auto-fix/<short-description>` (e.g. `auto-fix/missing-prisma-schema`)
- Read the problematic file, generate the minimal fix
- Commit with message: `fix: <description of what was fixed>`
- Open PR from fix branch into default branch
- **Never merge** — return PR URL for user approval

### 5. NOTIFY

Send to all configured notification channels:

**If fix PR created:**
```
Failure detected for [app name].
Root cause: [one-line summary].
Auto-fix PR: [pr_url]
Review and merge to trigger redeploy.
```

**If no fix possible:**
```
Failure detected for [app name].
Root cause: [one-line summary].
Recommended action: [specific next step].
```

**If diagnosis inconclusive:**
```
Issue detected for [app name].
Findings: [what was observed].
Unable to determine root cause automatically.
Please investigate: [specific things to check].
```

### 6. VERIFY (after user merges fix)

If the fix PR is merged and a new deployment triggers:

- Run post-deploy-verification checks
- If healthy: notify "Issue resolved after fix merge"
- If still failing: escalate — "Fix did not resolve the issue, further investigation needed"

## Rules

- Never merge PRs automatically — always require user approval
- Never push to main/master — always use fix branches
- Do not retry the same fix more than once
- Maximum 3 auto-fix attempts per incident before escalating to user
- Include all relevant context when delegating to sub-agents
- Every response must end with a concrete result or completed action

## Anti-Patterns

- **Fixing symptoms instead of root cause**: If the container OOM-kills, don't just increase memory — investigate the memory consumer
- **Auto-fixing infrastructure issues**: Server-level problems (disk full, Docker daemon down) can't be fixed via code PR
- **Notifying without actionable information**: "Something went wrong" is useless — always include what failed, why, and what to do
- **Cascading fixes**: If fix A causes failure B, stop and escalate — don't chain auto-fixes

## Related Skills

- **`failure-diagnosis`** — Pattern tables for identifying root causes
- **`rollback-strategy`** — When to rollback vs fix forward
- **`post-deploy-verification`** — Verify fix worked after merge

## Event Context

Your prompt contains the full incident context formatted by the event pipeline. This includes the event type, source details, error information, and any relevant identifiers (application, deployment, repository, etc.). Use all provided context to drive your investigation.

## Safety Rules

- Never merge PRs. Always return the PR URL for user approval.
- Never push to main/master. Always create a fix branch.
- If you cannot determine the root cause, notify the user with what you found and stop.
- Do not retry the same fix more than once. Maximum 3 auto-fix attempts per incident before escalating.
- Include all relevant context identifiers when delegating to diagnostics or github.
- After delegation returns, immediately process the result. Never say work is "underway".
- Every response must end with concrete information or a completed action.

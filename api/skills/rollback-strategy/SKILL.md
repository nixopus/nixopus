---
name: rollback-strategy
description: Guide rollback decisions after failed deployments — when to rollback vs retry, verification after rollback, and state preservation. Use when a deployment fails repeatedly, the app is unreachable after deploy, or the user requests a rollback.
metadata:
  version: "1.0"
---

# Rollback Strategy

## When to Rollback

| Situation | Action | Reason |
|---|---|---|
| Build failed, fix is obvious (typo, missing dep) | Retry with fix | Faster than rollback + fix + redeploy |
| Build failed 3+ times | Rollback | Deeper issue needs investigation without production down |
| Container crash loop after deploy | Rollback | Previous version was stable; fix in development |
| App unreachable after deploy (port/proxy issue) | Investigate first | May be config issue fixable without rollback |
| Database migration failed | Rollback with caution | See migration rollback section |
| User explicitly requests rollback | Rollback | User decision takes priority |

## Rollback Procedure

### 1. Preserve state before rolling back

Before executing rollback, capture:

- `get_deployment_logs` for the failed deployment — save the deployment ID and error
- `get_container_logs` if the container existed — last 100 lines
- `get_application` to note current env vars and config
- Record what changed between the working version and the failed one

### 2. Execute rollback

- `rollback_deployment` with the application ID
- This redeploys the last successful deployment's image

### 3. Verify rollback succeeded

After rollback completes, run the post-deploy-verification checks:

- Container is running and not restart-looping
- App is reachable internally and externally
- Healthcheck endpoint returns healthy
- No error patterns in logs

If verification fails after rollback: the previous version may also be broken (database schema change, expired credentials, infrastructure issue). Escalate to user.

### 4. Report to user

Include in the rollback report:

- What was deployed (commit/image that failed)
- Why it failed (root cause from diagnosis)
- What was rolled back to (previous deployment ID)
- Current status (verified healthy or still unhealthy)
- Recommended next steps (fix the issue, then redeploy)

## Database Migration Rollback

Migrations make rollbacks risky because the database schema may have changed:

| Migration type | Rollback safe? | Action |
|---|---|---|
| Additive only (new columns, new tables) | Yes | Old code ignores new columns |
| Column rename or removal | No | Old code references old column name |
| Data transformation | No | Transformed data may not work with old code |
| Index changes only | Yes | Old code unaffected |

If the migration is not rollback-safe:

1. Do NOT auto-rollback — notify user with the risk
2. Suggest: fix forward (deploy a corrected version) rather than rollback
3. If rollback is essential: include the reverse migration in the rollback

## Anti-Patterns

- **Rolling back repeatedly without investigating**: Each rollback buys time but doesn't fix the issue
- **Rolling back past multiple versions**: Only rollback to the immediately previous successful deployment
- **Rolling back infrastructure changes**: If the failure is DNS, proxy, or server-level, rollback won't help — the app code didn't change
- **Rolling back during active database migration**: Can corrupt data if migration is partially applied

## Related Skills

- **`post-deploy-verification`** — Run after rollback to confirm the previous version is healthy
- **`failure-diagnosis`** — Diagnose the root cause before deciding to rollback
- **`database-migration`** — Safe migration strategies that reduce rollback risk

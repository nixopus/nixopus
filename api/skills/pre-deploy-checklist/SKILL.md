---
name: pre-deploy-checklist
description: Validate deployment readiness before triggering a build — check Dockerfile, ports, env vars, healthchecks, and resource config. Use before any deployment to catch common configuration issues early.
metadata:
  version: "1.0"
---

# Pre-Deploy Checklist

Run through this checklist before triggering any deployment. Each check uses workspace tools. Report all findings, do not stop at the first failure.

## Checklist

### 1. Dockerfile exists and is valid

- Use `read_file` on the Dockerfile path
- Verify it has a `FROM` directive
- Verify it has an `EXPOSE` directive matching the expected port
- Verify it ends with a `CMD` or `ENTRYPOINT`
- If using multi-stage, verify the final stage copies the built artifacts

**If Dockerfile is missing**: Use the `dockerfile-generation` skill to generate one.

### 2. Port configuration matches

- Compare: Dockerfile `EXPOSE` value, app's actual listen port, any `PORT` env var, and the port configured in the Nixopus application
- All must agree. Mismatched ports are a top deployment failure cause.
- If using docker-compose, also check the `ports:` mapping

### 3. Required env vars are set

- Run env detection (use `env-detection` skill) to find all required vars
- Cross-reference with what's configured in the Nixopus application
- Flag any missing required vars
- Flag any vars using placeholder values (`your-api-key-here`, `change-me`, `TODO`)

### 4. Build command works

- Check `package.json` `scripts.build` (or equivalent) exists
- If TypeScript, check `tsconfig.json` exists and `outDir` is set
- Check that the build output directory referenced in the Dockerfile matches the actual build output

### 5. Dependencies are locked

- Check for a lockfile (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `Cargo.lock`, `poetry.lock`, `go.sum`)
- Using `npm install` instead of `npm ci` in a Dockerfile without a lockfile leads to inconsistent builds

### 6. .dockerignore exists

- Check for `.dockerignore` file
- Must include at minimum: `node_modules`, `.git`, `dist`, `.env`
- Missing `.dockerignore` causes bloated build contexts and potential secret leaks

### 7. Healthcheck endpoint

- For API servers: check if there's a `/health` or `/healthz` or `/api/health` endpoint
- If the Dockerfile or compose file includes a `HEALTHCHECK`, verify the endpoint exists in code
- Not strictly required but recommended — flag as warning if missing

### 8. Database migrations

- Check if the app has a migration system (`prisma`, `typeorm`, `knex`, `alembic`, `django migrate`, `goose`)
- If yes, verify the migration command is included in the deployment flow (compose `command:`, Dockerfile `CMD`, or Nixopus pre-deploy hook)
- Unmigrated databases after deploy cause runtime crashes

## Result format

Report as a table:

| Check | Status | Details |
|-------|--------|---------|
| Dockerfile | PASS/FAIL/WARN | What was found or missing |
| Port match | PASS/FAIL | Expected vs actual |
| Env vars | PASS/FAIL | Count of missing vars |
| Build command | PASS/FAIL | The command found |
| Lockfile | PASS/WARN | Which lockfile, or none |
| .dockerignore | PASS/WARN | Present or missing |
| Healthcheck | PASS/WARN | Endpoint found or none |
| Migrations | PASS/WARN/N/A | Migration tool and command |

Only block deployment (report FAIL) for checks 1-4. Checks 5-8 are warnings that should be reported but don't block.

## Summary Format
Report the checklist table, then:
**Ready**: what looks good
**Warnings**: non-critical issues
**Blockers**: must fix before deploy
**Recommendations**: specific fixes with code blocks

---
name: api-catalog
description: Reference for all Nixopus API operations callable via nixopus_api(method, path, body)
---

[api-catalog]
Use `nixopus_api({ method, path, body? })` for ALL Nixopus API calls below.
Pass the HTTP method, the full path (embed path params and query strings directly in the string), and an optional body object for POST/PUT/PATCH/DELETE payloads.

CALLING FORMAT:
- `method` — one of: GET, POST, PUT, PATCH, DELETE
- `path`   — full API path with path params and query strings embedded as a string
- `body`   — JSON object for request body (POST/PUT/PATCH/DELETE with payloads)

EXAMPLES:
```
nixopus_api({ method: "GET",  path: "/api/v1/deploy/applications?page=1&page_size=10" })
nixopus_api({ method: "GET",  path: "/api/v1/deploy/application?id=APP_UUID" })
nixopus_api({ method: "GET",  path: "/api/v1/deploy/application/deployments?id=APP_UUID&limit=5" })
nixopus_api({ method: "GET",  path: "/api/v1/deploy/application/deployments/DEPLOY_UUID" })
nixopus_api({ method: "GET",  path: "/api/v1/deploy/application/deployments/DEPLOY_UUID/logs?page=1&page_size=50" })
nixopus_api({ method: "POST", path: "/api/v1/deploy/application/restart", body: { id: "DEPLOY_UUID" } })
nixopus_api({ method: "POST", path: "/api/v1/deploy/application/redeploy", body: { id: "APP_UUID" } })
nixopus_api({ method: "GET",  path: "/api/v1/machines/stats" })
nixopus_api({ method: "POST", path: "/api/v1/notification/send", body: { channel: "slack", message: "deploy done" } })
```

COMMON MISTAKES TO AVOID:
- ❌ DO NOT use `operation` or `params` fields — the tool does not accept those
- ❌ DO NOT use `application_id` as the query param for deployments — use `id`
- ❌ DO NOT use GET for mutations (restart, redeploy, rollback) — they are POST
- ❌ DO NOT put path params as body fields — embed them in the path string
- ❌ DO NOT include `deploy_on_create` in any request body — it is an unknown field and the API rejects the entire request with HTTP 400.
- ❌ DO NOT omit `environment` from `POST /api/v1/deploy/application` — it is required. Valid values: `production`, `staging`, `development`.
- ❌ DO NOT omit `build_pack` from `POST /api/v1/deploy/application` — it is required. Valid values: `dockerfile`, `nixpacks`, `static`.
- ❌ DO NOT omit `port` from `POST /api/v1/deploy/application` — it is required (1–65535).
- ❌ DO NOT use `source: "github_public"` — it is not a valid value and will be rejected.
- ✅ `source` valid values: `github` (default — requires connector), `public_git` (any public HTTPS URL), `s3`, `zip`, `template`.
- ✅ `repository` format depends on `source`:
  - `source: "github"` (or omitted) → numeric GitHub connector repo ID. Call `GET /api/v1/github-connector/repositories`, use the integer `id`. NEVER pass owner/repo slug.
  - `source: "public_git"` → full HTTPS git URL, e.g. `"https://github.com/nixopus/sample-app.git"`. Use this when the repo is NOT in the GitHub connector.
- ✅ For `get_application_deployments` → `GET /api/v1/deploy/application/deployments?id=APP_UUID` (NOT `?application_id=`)
- ✅ For `get_deployment_by_id` → `GET /api/v1/deploy/application/deployments/DEPLOY_UUID` (path param, NOT query)
- ✅ For `restart_deployment` → `POST /api/v1/deploy/application/restart` with `body: { id: "DEPLOY_UUID" }`
- ✅ Container ops use path param: `GET /api/v1/container/CONTAINER_ID` (NOT `?id=`)

REQUIRED FLOW FOR DEPLOYMENT:
1. Check if repo is in `GET /api/v1/github-connector/repositories`
   - If YES → use `source: "github"`, `repository: <numeric_id>`
   - If NO  → use `source: "public_git"`, `repository: "https://github.com/org/repo.git"`
2. `POST /api/v1/deploy/application/project` with `{ name, repository, source, branch, ... }`
3. `POST /api/v1/deploy/application/project/deploy` with `{ id: <project_uuid> }`
4. Poll `GET /api/v1/deploy/application/deployments?id=<project_uuid>` until status is success/failed

## Applications
GET  /api/v1/deploy/applications                                         — List apps. Query: page?, page_size?, sort_by?, sort_direction?
GET  /api/v1/deploy/application?id={app_uuid}                           — Get one app by UUID
GET  /api/v1/deploy/application/deployments?id={app_uuid}               — List deployments. ⚠ param is `id` NOT `application_id`. Query: page?, limit?
GET  /api/v1/deploy/application/deployments/{deployment_id}             — Get deployment by UUID (path param, NOT query string)
GET  /api/v1/deploy/application/deployments/{deployment_id}/logs        — Deployment logs. Query: page?, page_size?, level?, start_time?, end_time?, search_term?
GET  /api/v1/deploy/application/logs/{application_id}                   — App-wide logs (path param)
POST /api/v1/deploy/application                                          — Create app. Body: { repository (INTEGER numeric GitHub repo ID — NOT "owner/repo"), name?, branch?, port?, build_pack?, dockerfile_path?, base_path?, environment_variables?, domains? }
PUT  /api/v1/deploy/application                                          — Update app config. Body: { id, ...fields }
DELETE /api/v1/deploy/application                                        — Delete app. Body: { id }
POST /api/v1/deploy/application/redeploy                                 — Rebuild & redeploy. Body: { id (app UUID), force?, force_without_cache? }
POST /api/v1/deploy/application/restart                                  — Restart deployment (no rebuild). Body: { id (deployment UUID) }
POST /api/v1/deploy/application/rollback                                 — Rollback. Body: { id (app UUID) }
POST /api/v1/deploy/application/cancel-deployment                        — Cancel in-flight. Body: { deployment_id }
POST /api/v1/deploy/application/recover                                  — Recover. Body: { application_id? } (omit to recover all)
PUT  /api/v1/deploy/application/labels?id={app_uuid}                    — Update labels. Body: (labels payload)
POST /api/v1/deploy/application/domains?id={app_uuid}                   — Add domain. Body: { domain, service_name?, port? }
DELETE /api/v1/deploy/application/domains?id={app_uuid}                 — Remove domain. Body: { domain }
GET  /api/v1/deploy/application/compose-services?id={app_uuid}          — List compose services
GET  /api/v1/deploy/application/servers?id={app_uuid}                   — Get server assignment
PUT  /api/v1/deploy/application/servers                                  — Set servers. Body: { application_id, server_ids, primary_server_id?, routing_strategy? }

## Projects
POST /api/v1/deploy/application/project                                  — Create project. Body: { name, repository (INTEGER numeric GitHub repo ID — NOT "owner/repo"), branch?, build_pack?, port?, environment?, ... }
POST /api/v1/deploy/application/project/deploy                           — Deploy project. Body: { id }
GET  /api/v1/deploy/application/project/family?family_id={id}            — Get project family
GET  /api/v1/deploy/application/project/family/environments?family_id={id} — List family environments
POST /api/v1/deploy/application/project/add-to-family                   — Add project to family

## Deploy Artifacts
GET    /api/v1/deploy/artifacts?application_id={app_uuid}               — List artifacts for app (⚠ this one uses `application_id`)
GET    /api/v1/deploy/artifacts/{deployment_id}/download                 — Download URL (path param)
DELETE /api/v1/deploy/artifacts/{deployment_id}                          — Delete artifact (path param)

## Domains
GET    /api/v1/domain                                                    — List domains. Query: type?
GET    /api/v1/domain/generate                                           — Generate random subdomain
POST   /api/v1/domain/custom                                             — Add custom domain
DELETE /api/v1/domain/custom                                             — Remove custom domain
POST   /api/v1/domain/verify                                             — Verify domain
GET    /api/v1/domain/dns-check?id={custom_domain_id}                   — DNS check

## GitHub Connectors
POST   /api/v1/github-connector                                          — Create connector. Body: { app_id, client_id, client_secret, pem, slug, webhook_secret }
PUT    /api/v1/github-connector                                          — Update connector. Body: { connector_id, installation_id }
DELETE /api/v1/github-connector                                          — Delete connector. Body: { id }
GET    /api/v1/github-connector/all                                      — List connectors
GET    /api/v1/github-connector/repositories                             — List GitHub repos
POST   /api/v1/github-connector/repository/branches                     — List branches. Body: { repository_name }

## Containers
GET    /api/v1/container                                                 — List containers. Query: page?, page_size?, status?, search?
GET    /api/v1/container/{container_id}                                  — Get container (path param)
POST   /api/v1/container/{container_id}/logs                             — Logs (path param). Body: { id, follow?, tail?, since?, until? }
POST   /api/v1/container/{container_id}/start                            — Start (path param)
POST   /api/v1/container/{container_id}/stop                             — Stop (path param)
POST   /api/v1/container/{container_id}/restart                          — Restart (path param)
DELETE /api/v1/container/{container_id}                                  — Remove (path param)
PUT    /api/v1/container/{container_id}/resources                        — Update resources. Body: { cpu_shares?, memory?, memory_swap? }
POST   /api/v1/container/images                                          — List images. Body: { all?, container_id?, image_prefix? }

## Machines
GET    /api/v1/machines                                                  — List servers. Query: page?, page_size?, search?
GET    /api/v1/machines/stats                                            — Host stats (CPU/RAM/disk/network) — no params
POST   /api/v1/machines/exec                                             — ⚠ Run command on host. Body: { command }
GET    /api/v1/machines/ssh/status                                       — SSH status for all machines
GET    /api/v1/machines/{id}/ssh/status                                  — SSH status for one machine (path param)
PUT    /api/v1/machines/{id}/set-default                                 — Set as org default (path param)
GET    /api/v1/machines/status                                           — Lifecycle status
POST   /api/v1/machines/restart                                          — ⚠ Restart machine
POST   /api/v1/machines/pause                                            — ⚠ Pause machine
POST   /api/v1/machines/resume                                           — ⚠ Resume machine
GET    /api/v1/machines/metrics                                          — Time-series metrics. Query: from, to, limit
GET    /api/v1/machines/metrics/summary                                  — Summarized metrics. Query: from, to
GET    /api/v1/machines/events                                           — Lifecycle events. Query: from, to, limit
GET    /api/v1/machines/backup/schedule                                  — Get backup schedule
PUT    /api/v1/machines/backup/schedule                                  — Update backup schedule
GET    /api/v1/machines/backups                                          — List backups
POST   /api/v1/machines/backup                                           — ⚠ Trigger backup

## System
GET    /api/v1/health                                                    — Health check (no auth)
GET    /api/v1/update/check                                              — Check for updates
POST   /api/v1/update                                                    — ⚠ Trigger update
GET    /api/v1/audit/logs                                                — Audit logs. Query: page?, page_size?, search?, resource_type?
GET    /api/v1/feature-flags                                             — List feature flags
GET    /api/v1/feature-flags/check?feature_name={name}                  — Check one flag
PUT    /api/v1/feature-flags                                             — ⚠ Update feature flag

## MCP
GET    /api/v1/mcp/catalog                                               — List MCP provider catalog
GET    /api/v1/mcp/servers                                               — List org MCP servers
POST   /api/v1/mcp/servers                                               — Add MCP server
PUT    /api/v1/mcp/servers/{id}                                          — Update MCP server (path param)
DELETE /api/v1/mcp/servers                                               — ⚠ Delete MCP server. Body: { id }
POST   /api/v1/mcp/servers/test                                          — Test connection
GET    /api/v1/mcp/internal/tools                                        — Discover MCP tools
GET    /api/v1/mcp/internal/servers                                      — List enabled servers
POST   /api/v1/mcp/internal/tools/call                                   — Call MCP tool. Body: { server_id, tool_name, arguments? }

## Notifications
POST   /api/v1/notification/send                                         — Send notification. Body: { channel (slack|discord|email), message, subject?, to?, metadata? }
GET    /api/v1/notification/preferences                                  — Get preferences
PATCH  /api/v1/notification/preferences                                  — Update preferences
GET    /api/v1/notification/smtp?id={org_id}                             — Get SMTP config
POST   /api/v1/notification/smtp                                         — Create SMTP config
PUT    /api/v1/notification/smtp                                         — Update SMTP config
DELETE /api/v1/notification/smtp                                         — Delete SMTP config
GET    /api/v1/notification/webhook/{type}                               — Get webhook (path param)
POST   /api/v1/notification/webhook                                      — Create webhook
PUT    /api/v1/notification/webhook                                      — Update webhook
DELETE /api/v1/notification/webhook                                      — Delete webhook

## Health Checks
POST   /api/v1/healthcheck                                               — Create health check
GET    /api/v1/healthcheck?application_id={app_uuid}                    — Get health checks for app
PUT    /api/v1/healthcheck                                               — Update health check
DELETE /api/v1/healthcheck?application_id={app_uuid}                    — Delete for app
PATCH  /api/v1/healthcheck/toggle                                        — Toggle health check
GET    /api/v1/healthcheck/results?application_id={app_uuid}            — Results. Query: limit?, start_time?, end_time?
GET    /api/v1/healthcheck/stats?application_id={app_uuid}              — Stats. Query: period?

## Extensions
GET    /api/v1/extensions                                                — List extensions. Query: category?, search?, type?, sort_by?, page?, page_size?
GET    /api/v1/extensions/categories                                     — List categories
GET    /api/v1/extensions/{id}                                           — Get extension by ID (path param)
GET    /api/v1/extensions/by-extension-id/{extension_id}                — Get by extension ID (path param)
[/api-catalog]

<!-- BEGIN REQUIRED FIELDS (auto-generated) -->
## REQUIRED FIELDS — auto-generated from Go struct tags (do not edit manually)
<!-- Run `go run api/scripts/gen_skill_required_fields.go` to regenerate -->

These fields have `validate:"required"` and WILL cause HTTP 400 if omitted.

### AddApplicationToFamilyRequest
- `name` (string): Application name
- `repository` (string): Numeric GitHub repository ID (get from GET /github-connector/repositories). **→ MUST be numeric GitHub repo ID — call GET /api/v1/github-connector/repositories and use the integer `id`. NEVER pass owner/repo slug or a URL.**

### CancelDeploymentRequest
- `deployment_id` (uuid.UUID): Deployment ID to cancel

### CreateDeploymentRequest
- `name` (string): Application name
- `environment` (shared_types.Environment): Deployment environment **→ Valid values: "production", "staging", "development"**
- `build_pack` (shared_types.BuildPack): Build strategy for the application **→ Valid values: "dockerfile", "nixpacks", "static"**
- `repository` (string): Numeric GitHub repository ID (get from GET /github-connector/repositories). Do NOT pass owner/repo slug — it will fail at parse time. **→ MUST be numeric GitHub repo ID — call GET /api/v1/github-connector/repositories and use the integer `id`. NEVER pass owner/repo slug or a URL.**
- `branch` (string): Git branch to deploy
- `port` (int): Port the application listens on

### CreateProjectRequest
- `name` (string): Project name
- `repository` (string): Numeric GitHub repository ID (get from GET /github-connector/repositories). Do NOT pass owner/repo slug — it will fail at parse time. **→ MUST be numeric GitHub repo ID — call GET /api/v1/github-connector/repositories and use the integer `id`. NEVER pass owner/repo slug or a URL.**

### CreateTemplateDeploymentRequest
- `template_id` (string): Template ID to deploy from
- `name` (string): Name for the deployed application

### DeleteDeploymentRequest
- `id` (uuid.UUID): Application ID to delete

### DeployProjectRequest
- `id` (uuid.UUID): Project ID to deploy

### DuplicateProjectRequest
- `source_project_id` (uuid.UUID): ID of the project to duplicate
- `environment` (shared_types.Environment): Environment for the duplicated project **→ Valid values: "production", "staging", "development"**

### GetProjectFamilyRequest
- `family_id` (uuid.UUID): Project family ID

### IsDomainAlreadyTakenRequest
- `domain` (string): Domain to check for availability

### IsDomainValidRequest
- `domain` (string): Domain to validate

### IsNameAlreadyTakenRequest
- `name` (string): Application name to check for uniqueness

### IsPortAlreadyTakenRequest
- `port` (int): Port number to check for availability

### PreviewComposeRequest
- `repository` (string): Numeric GitHub repository ID (get from GET /github-connector/repositories). **→ MUST be numeric GitHub repo ID — call GET /api/v1/github-connector/repositories and use the integer `id`. NEVER pass owner/repo slug or a URL.**
- `branch` (string): Git branch to preview

### ReDeployApplicationRequest
- `id` (uuid.UUID): Application ID to redeploy

### RestartDeploymentRequest
- `id` (uuid.UUID): Application ID to restart

### RollbackDeploymentRequest
- `id` (uuid.UUID): Application ID to roll back

### SetApplicationServersRequest
- `application_id` (uuid.UUID): Application ID to assign servers to
- `server_ids` ([]uuid.UUID): Server IDs to assign to the application

<!-- END REQUIRED FIELDS (auto-generated) -->

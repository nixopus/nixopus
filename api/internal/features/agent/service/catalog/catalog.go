package catalog

// Catalog is injected as a system message so the agent knows
// all available Nixopus API operations it can call via the nixopus_api tool.
const Catalog = `[api-catalog]
Use nixopus_api(method, path, body) for ALL Nixopus API calls below.
Pass the HTTP method and API path directly. For path params, embed them in the path string.

Example: nixopus_api({ method: "GET", path: "/api/v1/deploy/applications?page=1&page_size=10" })
Example: nixopus_api({ method: "POST", path: "/api/v1/deploy/application", body: { repository: "owner/repo", name: "my-app", port: 3000 } })
Example: nixopus_api({ method: "GET", path: "/api/v1/deploy/application/deployments?id=APP_UUID&page=1&page_size=10" })

## Applications
GET /api/v1/deploy/applications — List apps. Query: page, page_size, sort_by, sort_direction
GET /api/v1/deploy/application?id={app_uuid} — Get one app
GET /api/v1/deploy/application/deployments?id={app_uuid} — List deployments for app. Query: id, page, page_size
GET /api/v1/deploy/application/deployments/{deployment_id} — Get deployment by ID
GET /api/v1/deploy/application/deployments/{deployment_id}/logs — Deployment logs. Query: page, page_size, level, start_time, end_time, search_term
GET /api/v1/deploy/application/logs/{application_id} — App-wide logs (same filters as deployment logs)
POST /api/v1/deploy/application — Create/deploy app. Body: repository (STRING!), source?, name?, branch?, port?, build_pack?, dockerfile_path?, base_path?, environment_variables?, build_variables?, domains?
POST /api/v1/deploy/application/template — Deploy from template. Body: (template deploy request)
PUT /api/v1/deploy/application — Update app config. Body: { id, ...fields }
DELETE /api/v1/deploy/application — Delete app. Body: { id }
POST /api/v1/deploy/application/redeploy — Rebuild & redeploy. Body: { id, force?, force_without_cache?, target_server_ids? }
POST /api/v1/deploy/application/restart — Restart deployment (no rebuild). Body: { id }
POST /api/v1/deploy/application/rollback — Rollback. Body: { id } (application UUID; optional target_server_ids)
POST /api/v1/deploy/application/cancel-deployment — Cancel in-flight deployment. Body: { deployment_id }
POST /api/v1/deploy/application/recover — Recover app(s). Body: { application_id? } (omit to recover all)
PUT /api/v1/deploy/application/labels?id={app_uuid} — Update labels. Body: (labels payload)
POST /api/v1/deploy/application/domains?id={app_uuid} — Add domain to app. Body: { domain, service_name?, port? }
DELETE /api/v1/deploy/application/domains?id={app_uuid} — Remove domain. Body: { domain }
GET /api/v1/deploy/application/compose-services?id={app_uuid} — List compose services
POST /api/v1/deploy/application/preview-compose — Preview compose. Body: (preview request)
GET /api/v1/deploy/application/servers?id={app_uuid} — Get app server assignment
PUT /api/v1/deploy/application/servers — Set app servers. Body: { application_id, server_ids, primary_server_id?, routing_strategy? }

## Projects
POST /api/v1/deploy/application/project — Create project. Body: { name, repository, ... }
POST /api/v1/deploy/application/project/deploy — Deploy project. Body: { id }
POST /api/v1/deploy/application/project/duplicate — Duplicate project
GET /api/v1/deploy/application/project/family — Get project family. Query: family_id
GET /api/v1/deploy/application/project/family/environments — List family environments. Query: family_id
POST /api/v1/deploy/application/project/add-to-family — Add project to family

## Deploy artifacts
GET /api/v1/deploy/artifacts?application_id={app_uuid} — List deployment artifacts for an app
GET /api/v1/deploy/artifacts/{deployment_id}/download — Presigned download URL for artifact
DELETE /api/v1/deploy/artifacts/{deployment_id} — Delete artifact

## Domains
GET /api/v1/domain — List domains. Query: type?
GET /api/v1/domain/generate — Generate random subdomain
POST /api/v1/domain/custom — Add custom domain
DELETE /api/v1/domain/custom — Remove custom domain
POST /api/v1/domain/verify — Verify custom domain
GET /api/v1/domain/dns-check?id={custom_domain_id} — Check DNS for custom domain

## GitHub Connectors
POST /api/v1/github-connector — Create connector. Body: { app_id, client_id, client_secret, pem, slug, webhook_secret }
PUT /api/v1/github-connector — Update connector. Body: { connector_id, installation_id }
DELETE /api/v1/github-connector — Delete connector. Body: { id }
GET /api/v1/github-connector/all — List connectors
GET /api/v1/github-connector/repositories — List GitHub repos
POST /api/v1/github-connector/repository/branches — List branches. Body: { repository_name }

## Containers
GET /api/v1/container — List containers. Query: page, page_size, sort_by, sort_order, search, status, name, image
GET /api/v1/container/{container_id} — Get container
POST /api/v1/container/{container_id}/logs — Container logs. Body: { id, follow?, tail?, since?, until?, stdout?, stderr? }
POST /api/v1/container/{container_id}/start — Start container
POST /api/v1/container/{container_id}/stop — Stop container
POST /api/v1/container/{container_id}/restart — Restart container
DELETE /api/v1/container/{container_id} — Remove container
PUT /api/v1/container/{container_id}/resources — Update resources. Body: cpu_shares?, memory?, memory_swap?
POST /api/v1/container/images — List images. Body: { all?, container_id?, image_prefix? }
POST /api/v1/container/prune/build-cache — Prune build cache
POST /api/v1/container/prune/images — Prune images

## Machines
GET /api/v1/machines — List servers. Query: page, page_size, search, sort_by, sort_order, status, is_active
POST /api/v1/machines — Register BYOS machine
POST /api/v1/machines/{id}/verify — Verify SSH for machine
PATCH /api/v1/machines/{id}/rename — Rename machine
DELETE /api/v1/machines/{id} — Remove machine
GET /api/v1/machines/ssh/status — SSH status for all machines
GET /api/v1/machines/{id}/ssh/status — SSH status for one machine
PUT /api/v1/machines/{id}/set-default — Set machine as org default
GET /api/v1/machines/stats — Host stats (CPU/RAM/disk/network)
POST /api/v1/machines/exec — Run command on host. Body: { command }
GET /api/v1/machines/status — Lifecycle status of provisioned machine
POST /api/v1/machines/restart — ⚠ Restart machine
POST /api/v1/machines/pause — ⚠ Pause machine
POST /api/v1/machines/resume — ⚠ Resume machine
GET /api/v1/machines/metrics — Time-series metrics. Query: from, to, limit
GET /api/v1/machines/metrics/summary — Summarized metrics. Query: from, to
GET /api/v1/machines/events — Lifecycle/events. Query: from, to, limit
GET /api/v1/machines/plans — List machine plans
POST /api/v1/machines/plan/select — Select plan (billing). Body: (plan selection payload)
GET /api/v1/machines/billing — Billing status for org machine
GET /api/v1/machines/backup/schedule — Get backup schedule
PUT /api/v1/machines/backup/schedule — Update backup schedule
GET /api/v1/machines/backups — List backups. Query: page, page_size, search, sort_by, sort_order, status
POST /api/v1/machines/backup — ⚠ Trigger backup now
POST /api/v1/machines/trial/provision — Provision trial resources
GET /api/v1/machines/trial/status/{sessionId} — Trial session status

## System
GET /api/v1/health — Public health check (no auth)
GET /api/v1/update/check — Check for updates
POST /api/v1/update — ⚠ Trigger system update
GET /api/v1/audit/logs — Audit logs. Query: page, page_size, search, resource_type
GET /api/v1/feature-flags — List feature flags
GET /api/v1/feature-flags/check?feature_name={name} — Check one flag
PUT /api/v1/feature-flags — ⚠ Update feature flag

## MCP
GET /api/v1/mcp/catalog — List MCP provider catalog
GET /api/v1/mcp/servers — List org MCP servers
POST /api/v1/mcp/servers — Add MCP server
PUT /api/v1/mcp/servers/{id} — Update MCP server
DELETE /api/v1/mcp/servers — ⚠ Delete MCP server. Body: { id }
POST /api/v1/mcp/servers/test — Test connection. Body: (test request)
GET /api/v1/mcp/internal/tools — Discover MCP tools
GET /api/v1/mcp/internal/servers — List enabled servers with credentials
POST /api/v1/mcp/internal/tools/call — Call MCP tool. Body: { server_id, tool_name, arguments? }

## Extensions
GET /api/v1/extensions — List extensions. Query: category, search, type, sort_by, sort_dir, page, page_size
GET /api/v1/extensions/categories — List categories
GET /api/v1/extensions/{id} — Get extension by ID
GET /api/v1/extensions/by-extension-id/{extension_id} — Get extension by extension ID

## Notifications
POST /api/v1/notification/send — Send notification. Body: { channel (slack|discord|email), message, subject?, to?, metadata? }
PATCH /api/v1/notification/preferences — Update preferences
GET /api/v1/notification/preferences — Get preferences
GET /api/v1/notification/smtp?id={org_id} — Get SMTP config
POST /api/v1/notification/smtp — Create SMTP config
PUT /api/v1/notification/smtp — Update SMTP config
DELETE /api/v1/notification/smtp — Delete SMTP config
GET /api/v1/notification/webhook/{type} — Get webhook config
POST /api/v1/notification/webhook — Create webhook
PUT /api/v1/notification/webhook — Update webhook
DELETE /api/v1/notification/webhook — Delete webhook

## Health checks
POST /api/v1/healthcheck — Create health check. Body: (create payload)
GET /api/v1/healthcheck?application_id={app_uuid} — Get health checks for application
PUT /api/v1/healthcheck — Update health check. Body: (update payload)
DELETE /api/v1/healthcheck?application_id={app_uuid} — Delete health check for application
PATCH /api/v1/healthcheck/toggle — Toggle health check. Body: (toggle payload)
GET /api/v1/healthcheck/results?application_id={app_uuid} — Results. Query: limit?, start_time?, end_time?
GET /api/v1/healthcheck/stats?application_id={app_uuid} — Stats. Query: period?
[/api-catalog]`

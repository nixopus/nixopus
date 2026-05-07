---
name: deploy-flow
description: Full deploy pipeline — source detection, hints-driven analysis, project creation, deployment monitoring, and live URL delivery. Load when the user wants to deploy an application.
metadata:
  version: "2.0"
---

# Deploy Flow

## Sample App Fast Path
If a `[sample-app-fast-path]` block is injected in the system messages, follow it exactly. Do NOT load this skill or any other skill — the fast path already contains the complete recipe.

## Source Detection — FIRST STEP
Check the user message for a [context] block before calling any tool.

**Frontend context block** — `[Context: Repository "owner/repo" (Language: ..., Branch: ..., Visibility: public|private)]`:
- **Visibility: public** → first call `load_tool({ toolName: "load_remote_repository" })` (it is a searchable tool), then call `load_remote_repository` with `repoUrl: "https://github.com/owner/repo"` and the branch. Do NOT call `get_github_repositories` or `analyze_repository` — `load_remote_repository` already returns hints. Then continue the hints-driven deploy flow (step 2 below). Source and repository are set automatically.
- **Visibility: private** → the repo must be in connected GitHub repos. Check `[user-context]` for the numeric repo ID. If found, use the **standard GitHub flow** (routing step 1). If not found, ask the user to connect the repo.

**source=s3** — The context block contains syncTarget (the S3 storage ID for files). It may also contain applicationId if the app already exists.
- Call load_local_workspace with the syncTarget value. It returns `hints` with ecosystem, framework, port, Dockerfile presence, and confidence levels.
- If applicationId is present: the app exists. Use it for deploy_project (or quick_deploy if redeploying from scratch).
- If applicationId is absent: this is a first deploy. Use the returned hints to decide config, then quick_deploy (or createProject if you need to set env vars first). Source and repository are set automatically for S3.
- GitHub connector tools are blocked for S3 sources — do not attempt them.

**No context block** — Check `[user-context]` first — it has connectors, repos, apps, servers, and domains pre-loaded. Route from connector access and what the user asked for; do **not** treat missing GitHub connector as a hard stop if a public git URL can be used instead.

**Behavior checklist (routing expectations):**

- **Case A:** No connector + user provides a valid public HTTPS git URL → call `load_remote_repository` — it returns hints. Continue the hints-driven deploy flow.
- **Case B:** Connectors/repos available + user includes a URL in the message → default to the **standard GitHub flow** below (use connected repo access), unless the user **explicitly** asked to deploy from a pasted public URL instead of connected repos. **Explicit URL preference** means wording that clearly bypasses the connector for that deploy, for example: "deploy from this URL", "clone `https://github.com/org/repo.git` — don't use my connected repos", "ignore my GitHub connection and use this public repo link", "use the pasted HTTPS link, not my org's connected repositories".
- **Case C:** No connector + no valid public HTTPS git URL in the message → ask the user for a public HTTPS clone URL; you may optionally point them at `read_skill("github-onboarding")` as an alternative — onboarding is **not** a hard blocker.

**Routing (no `[context]` block, no frontend context block):**

1. **Connectors/repos available** (per `[user-context]`, or refresh with nixopus_api('get_github_repositories') / nixopus_api('get_applications') only if context is clearly stale — e.g. user says they just connected GitHub but `[user-context]` still shows no connectors; repo/app lists contradict what a fresh call returns; or the user reports an error about a resource that does not appear in context) **and** the user did **not** explicitly request deploying from a public git URL (connector bypass — see **Case B** examples above): **standard GitHub flow.** Use pre-loaded connector/repo IDs instead of redundant API calls when possible. Then: `analyze_repository` (returns hints) → use hints to decide config → `quick_deploy` or `createProject` + nixopus_api('deploy_project', { id }).

   **Fallback (connector path failed, URL still viable):** If you already chose the standard GitHub flow but a required GitHub/connector step fails (auth, missing repo, permission, transient API error, etc.) **and** the user's message still contains a **valid public HTTPS git clone URL** you can use: switch to **`load_remote_repository`** with that URL (and branch if given), then continue the **hints-driven** deploy flow. Prefer this over repeatedly retrying the connector when the public URL is a known-good alternative. Connector-first ordering is unchanged: attempt the connector path first when the rules above apply; only fall back after a real failure.

2. **Connector/repo access missing** and the user provided a **valid public HTTPS git clone URL** (e.g. `https://github.com/org/repo.git` or `https://…/org/repo` — not `git@…` SCP-style, not non-HTTPS schemes): call **`load_remote_repository`** with that URL (and branch if the user specified one) — it returns hints. Then continue the **hints-driven** deploy flow: use hints to decide config, then `quick_deploy` or `createProject` / `deploy_project`.

3. **Connector/repo access missing** and **no** valid public HTTPS git URL is present: **ask** the user for a public HTTPS clone URL they can share. Optionally mention `read_skill("github-onboarding")` if they prefer to connect GitHub instead — do **not** refuse to continue the deploy solely because onboarding was skipped.

## Deploy Patterns
A [deploy-patterns] block may be injected with known fixes, pitfalls, and fast paths for the detected ecosystem. When present, check it before diagnosing issues — if a known fix matches, apply it directly instead of re-investigating from scratch.

## Deploy Steps
1. **Load codebase** — call ONE of: analyze_repository, load_remote_repository, or load_local_workspace. All three return a **hints** object. NEVER call more than one for the same repo — they are alternatives, not complements.
   - `load_remote_repository` → public repos (loaded via URL). Returns hints + loads workspace for S3 deploy.
   - `analyze_repository` → connected GitHub repos only (requires numeric repo ID from connectors).
   - `load_local_workspace` → S3/synced sources (requires syncTarget from context block).

2. **Use hints to decide next action:**
   - **hints.confidence = "high"** — all fields are certain. Skip manual file exploration entirely. Proceed directly to step 6 (quick_deploy or createProject).
   - **hints.confidence = "medium"** or **warnings present** — verify only the flagged items with 1-2 targeted read_file calls (e.g. read package.json to confirm primary framework). Then proceed to step 6.
   - **hints.confidence = "low"** on port or framework — use workspace tools (read_file, list_directory, grep) for those specific fields only. Binary fields (hasDockerfile, hasDockerCompose, packageManager, isMonorepo) are always certain and need no verification.

3. If hints.isMonorepo is true: read_skill("monorepo-strategy") for service discovery, dependency ordering, and build context strategy.

4. If hints.hasDockerfile is false: read_skill("dockerfile-generation") and the matching ecosystem skill (e.g. read_skill("node-deploy")). Also read_skill("dockerignore-generation") if hints.hasDockerignore is false. For static sites needing Caddy config, read_skill("caddyfile-generation").
   - **Workspace-backed sources** (`source=s3` after `load_local_workspace`, **or** codebase loaded via **`load_remote_repository`** from a public HTTPS git URL): Use **`write_workspace_files`** to save generated Dockerfile/.dockerignore/etc.
   - **GitHub connector flow** (standard GitHub flow): Do **not** use `write_workspace_files`. Use the GitHub tools directly: read_skill("github-workflow") and follow the Fix-via-PR flow (create branch → github_create_or_update_file on that branch → open PR). NEVER push directly to main/master.

5. If the app has database migrations (Prisma, TypeORM, Django, Alembic, etc.): read_skill("database-migration").

6. **Create and deploy** (if app doesn't exist):
   - **Preferred: quick_deploy** — creates the project and deploys in one step. Auto-generates a subdomain if domains is empty. Use this for first-time deploys when you have all the deployment config from hints. For compose: pass compose_services and compose_domains.
   - **Alternative: createProject then nixopus_api('deploy_project', { id })** — use when you need to modify the app between creation and deployment (e.g. setting env vars via nixopus_api('update_application', { id, ... })). createProject also auto-generates a subdomain if domains is empty.
   - For custom domains, read_skill("domain-attachment").

7. Monitor deployment — mandatory but lean. Call nixopus_api('get_application_deployments', { id, limit: 1 }) once to get the deployment ID. Then poll nixopus_api('get_deployment_by_id', { deployment_id }) only — do NOT call get_deployment_logs unless the status is failed/error or the user asks. One poll is enough if the build is fast; for slow builds, poll at most 2-3 times. Do NOT call get_application after deploy — you already have the app details from creation.
8. Verify: read_skill("post-deploy-verification") and run the verification checklist.
9. Share the live URL clearly and explicitly once the app is verified reachable.

## Rules
- Use createProject or quick_deploy to create apps. create_application does not exist.
- **repository parameter**: MUST be a numeric GitHub repo ID (e.g. "1203828551"), NEVER a slug like "owner/repo". Get it from `get_github_repositories` or `[user-context]` repos. When deploying via `load_remote_repository` (public URL flow), do NOT pass `repository` at all — source and repository are set automatically by the system.
- Check [user-context] apps (or nixopus_api('get_applications') if context is stale — use the same "clearly stale" signals as **Routing** step 1) before creating to avoid duplicates.
- Never hardcode secrets. Use nixopus_api('update_application', { id, environment_variables }) for env vars.
- Keep user-facing responses product-level. Do **not** mention internal implementation details like source storage internals, sync internals, tool names/IDs, hints, confidence levels, or context block fields.
- If the user asks to deploy, do not stop at planning, explanation, or diagnosis. Execute the deployment flow unless blocked by missing credentials, missing access, or required secrets.
- "Would you like me to fix this?" is a failure mode when the fix is obvious. Fix it.
- If an operation is async, keep polling and keep the user updated until there is a terminal outcome. Do not abandon the flow with promises of future follow-up.
- If you create a PR, include the URL in your reply. If it failed, say what failed.

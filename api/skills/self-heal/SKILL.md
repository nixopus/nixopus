---
name: self-heal
description: Self-healing loop for failed deployments — diagnose, fix, redeploy up to 3 attempts, then escalate or rollback. Load when a deployment fails or build errors occur.
metadata:
  version: "1.0"
---

# Self-Heal

## Flow (max 3 attempts)
On build_failed: check [deploy-patterns] for a known fix first → if no match, get_deployment_logs → diagnose → apply fix → redeploy → resume monitoring.
Known fixes from [deploy-patterns] have cross-org confidence scores. Prefer high-confidence fixes (>70%) before investigating from scratch.

### Applying the fix
- **S3 sources**: Use write_workspace_files to save the fix — files sync automatically.
- **GitHub sources**: Do NOT use write_workspace_files (it only writes locally and the fix will not reach the repo). Instead, use the GitHub tools: read_skill("github-workflow") and follow the Fix-via-PR flow (create branch → github_create_or_update_file → open PR → ask user to merge → redeploy).

### Rules
- Do not stop to ask the user unless the fix is ambiguous or requires credentials you do not have.
- After each failed attempt, tell the user what broke and what you are trying next.
- Maximum 3 self-heal attempts.
- After 3 failures, read_skill("rollback-strategy") to decide whether to rollback or escalate.
- If escalation is required, tell the user plainly that you could not complete the deployment automatically and that a Nixopus team member will reach out shortly to help finish it.

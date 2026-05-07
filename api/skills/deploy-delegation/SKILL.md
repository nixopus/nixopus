---
name: deploy-delegation
description: Sub-agent routing table — which agent handles diagnostics, machine health, infrastructure, GitHub, billing, and notifications. Load when the current task is not a direct deployment.
metadata:
  version: "1.1"
---

# Delegation

Use the `delegate` tool to route non-deploy tasks to specialized agents. Pass the agent name and a task description with all relevant context.

## Routing Table

| Agent | Handles | Example tasks |
|-------|---------|---------------|
| `diagnostics` | Build errors, crashes, runtime issues | "Investigate why deployment X failed" |
| `machine` | Server health, CPU/RAM, Docker daemon, DNS, backups | "Check server memory usage" |
| `infrastructure` | Domain listing/creation/deletion, containers, healthchecks, server management | "List all domains and their status" |
| `github` | Branches, PRs, file operations | "Create a fix branch and PR for the Dockerfile" |
| `preDeploy` | First-time validation, monorepo assessment | "Run pre-deploy checks on this repository" |
| `notification` | Deploy alerts, channel config | "Send a deploy success notification to Slack" |
| `billing` | Credits, plans, invoices | "Check credit balance" |

## Usage

```
delegate({ agent: "diagnostics", task: "Investigate deployment failure for applicationId=abc-123. Check logs and container state." })
```

Always include relevant identifiers in the task: applicationId, owner, repo, branch. The delegate tool automatically injects context formatting for agents that need it.

Delegation is synchronous — process the result in the same response. If delegation returns an error, try using direct tools instead.

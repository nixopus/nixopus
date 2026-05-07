---
name: diagnostic-workflow
description: Layer-by-layer diagnostic workflow for application and container issues — deployment logs, container state, HTTP probes. Load when investigating a deployment failure or runtime issue.
metadata:
  version: "1.0"
---

# Diagnostic Workflow

**All tools below are runtime tools. Use `search_tools` / `load_tool` to find and load them — do NOT use `skill_read` or `skill_search`.**

## Diagnostic Layers (IN ORDER, stop on root cause)
1. `get_application_deployments` — check deployment history and status
2. `get_deployment_logs` — read build and deploy logs for errors
3. `list_containers` → `search_tools("container logs")` → `load_tool(...)` → load needed tools
4. `get_container_logs` — check container runtime output
5. `search_tools("http probe")` → `load_tool("http_probe")` → probe public URL

If the issue appears application-level, check logs layer by layer. For container-level resource issues, defer to the Machine Agent which has host_exec.

If the issue appears to be server-level (CPU, RAM, disk, Docker daemon, DNS, proxy, or domain/TLS), defer to the Machine Agent.

Match log output against the pattern tables in the failure-diagnosis skill before hypothesizing. Tool 404 → skip layer. Root cause: bold summary, evidence in code block, fix in 1-2 sentences.

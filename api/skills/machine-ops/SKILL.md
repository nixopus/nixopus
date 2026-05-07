---
name: machine-ops
description: Machine-level diagnostic layers, lifecycle management (restart/pause/resume), metrics analysis, and backup operations. Load when investigating server health or managing machine state.
metadata:
  version: "1.0"
---

# Machine Operations

## Lifecycle Management
You can check and control the machine instance state:
- get_machine_lifecycle_status → current state (Running, Paused, Stopped), PID, uptime
- restart_machine → restart the instance (requires user approval)
- pause_machine → pause the instance (requires user approval)
- resume_machine → resume a paused instance (requires user approval)

Always check get_machine_lifecycle_status before performing restart/pause/resume.

## Metrics & Events
- get_machine_metrics → historical time-series metrics (CPU, memory, disk, network)
- get_machine_metrics_summary → summarized averages, peaks, and trends
- get_machine_events → lifecycle events (restarts, failures, state changes)

Use metrics for trend analysis and incident correlation. Use get_machine_stats for a point-in-time snapshot.

## Backups
- get_backup_schedule → current backup schedule configuration
- update_backup_schedule → modify backup frequency, retention, timing
- list_machine_backups → list available backups with timestamps and status
- trigger_machine_backup → create an immediate backup (requires approval)

## Diagnostic Layers (IN ORDER, stop on root cause)
1. get_servers_ssh_status → reachable?
2. get_machine_stats → CPU, RAM, disk, load, uptime
3. Anomalies: mem>90% → host_exec "ps aux --sort=-%mem | head -20". disk>85% → "du -sh /var/lib/docker/* 2>/dev/null | sort -rh | head -10". cpu>80% → "ps aux --sort=-%cpu | head -20". load>2x cores → overloaded.
4. Docker → host_exec "systemctl status docker --no-pager", "docker info 2>&1 | head -30"
5. System logs → host_exec "dmesg | tail -30", "journalctl -u docker --since '30 min ago' --no-pager | tail -50"
6. Proxy/domain: follow domain-tls-routing skill. Caddy status/logs/validate via host_exec. For domain CRUD or reachability checks, defer to Infrastructure Agent.
7. Network → host_exec "ss -tlnp"
8. Cleanup → host_exec "docker system df"

Root cause: bold summary, evidence in code block, fix in 1-2 sentences.
No anomalies: report healthy with key metrics.

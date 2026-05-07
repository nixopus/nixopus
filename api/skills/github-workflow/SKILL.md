---
name: github-workflow
description: Fix-via-PR workflow, file operations, connector resolution, and GitHub safety rules. Load when performing GitHub operations like creating branches, PRs, or file changes.
metadata:
  version: "1.0"
---

# GitHub Workflow

## Connector Resolution
When a connectorId is provided in the delegation message, use that connector_id when calling get_github_repositories to list repos from the correct GitHub account. If no connectorId is provided and there are multiple connectors, use get_github_connectors to list them and pick the first one with valid credentials, then use its ID for get_github_repositories.

## File Write Capabilities
- github_create_or_update_file: Create or update a single file. To update, first read the file with github_get_repo_file to get its current sha, then pass that sha. To create a new file, omit sha.
- github_create_branch: Create a new branch from a commit SHA. Use get_github_repository_branches to find the source branch HEAD SHA.
- github_create_pull_request: Open a PR from a head branch into a base branch.

## Fix-via-PR Flow
When asked to fix a file in a repo:
1. Call github_get_branch with the default branch name (e.g. "main") to get its HEAD commit SHA.
2. Create a fix branch with github_create_branch using that SHA (e.g. branch name "nixopus/fix-dockerfile").
3. Read the file to fix with github_get_repo_file on the default branch to get its content and blob sha.
4. Write the fixed file with github_create_or_update_file targeting the fix branch, passing the blob sha from step 3.
5. Create a PR with github_create_pull_request from the fix branch into the default branch.
6. Ask the user to review and merge the PR. Once merged, the changes land on the default branch.
7. After the PR is merged, call redeploy_application to pick up the changes from the default branch.
Return the PR URL, PR number, and fix branch name to the parent agent in your final message. Never say work is "underway" or that you will send the link later.

## Branch Limitation
The application's branch is set at creation time via create_application and CANNOT be changed via update_application. The app always deploys from its configured branch (usually main/master). This is why you MUST use the branch → PR → merge → redeploy flow: changes go to a feature branch, get merged into the default branch via PR, then redeploy picks them up.

## GitHub Safety — NON-NEGOTIABLE
- NEVER commit or push directly to main/master. The tool will reject it. Always create a feature branch → commit there → open a PR.
- Never merge PRs unless user explicitly requests. Return PR URL.
- No destructive ops (force push, branch delete, PR close) without user approval.

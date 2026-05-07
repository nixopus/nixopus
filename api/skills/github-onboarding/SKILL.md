---
name: github-onboarding
description: Guide users through connecting GitHub to Nixopus when no GitHub connector exists. Covers GitHub App installation for cloud users and GitHub App manifest setup for self-hosted users.
metadata:
  version: "1.0"
---

# GitHub Onboarding

Use this when `get_github_connectors` returns empty or has no valid connectors. Do NOT continue the deploy flow or call `get_github_repositories` until the user has completed the GitHub connection.

## Cloud Users

Guide the user step by step:

1. Install the Nixopus GitHub App: https://github.com/apps/nixopus/installations/new
   - For an organization: https://github.com/organizations/{org-name}/settings/apps/nixopus/installations/new
2. Select which repositories or the entire organization to grant access to.
3. After clicking Install, GitHub redirects back to Nixopus and the connection is saved automatically.
4. Once connected, come back and say "deploy my app" — Nixopus will list the repos and walk through deployment.

## Self-Hosted Users

Self-hosted users need to create their own GitHub App first:

1. Open the Nixopus dashboard and go to the Apps page.
2. The setup wizard will walk through creating a GitHub App using GitHub's app manifest flow.
3. Choose whether to create the app under a personal account or an organization.
4. After creating the app, install it on the desired repositories.
5. Once setup completes, come back and say "deploy my app" — Nixopus will list the repos and walk through deployment.

## Important

- Do NOT attempt `get_github_repositories`, `analyze_repository`, or any deploy steps until the connection is set up.
- NEVER ask the user whether they are on cloud or self-hosted. The [user-context] block always contains `instance: mode=cloud` or `instance: mode=self-hosted`. Use that directly. If mode is "cloud", give cloud steps. If mode is "self-hosted", give self-hosted steps.
- After the user says they've connected, call `get_github_connectors` again to verify before proceeding.

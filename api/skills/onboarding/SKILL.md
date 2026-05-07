---
name: onboarding
description: Proactive welcome flow for new Nixopus users. Triggered by the frontend __ONBOARD__ signal when a user has no GitHub connectors and no deployed applications. Greets the user warmly, explains what Nixopus is, and immediately loads the github-onboarding skill to deliver correct GitHub connection steps.
metadata:
  version: "1.1"
---

# Onboarding

This skill is triggered when you receive a message containing exactly `__ONBOARD__`.

## What to do

**Step 1 — Load the github-onboarding skill immediately:**

```
read_skill("github-onboarding")
```

You MUST call this before composing your response. The github-onboarding skill contains the correct, authoritative steps for connecting GitHub (including the right links and flows for cloud vs self-hosted). Do not guess or invent navigation paths.

**Step 2 — Compose a single welcome response that includes:**

1. A brief greeting (one sentence). Plain text only, no emojis, no markdown headers.
2. One sentence explaining what Nixopus is: a platform where you describe what you want to deploy and it handles the rest, delivering a live URL.
3. The GitHub connection steps from the github-onboarding skill — use the correct flow for the user's instance type:
   - **Cloud (managed Nixopus):** Direct them to install the GitHub App at the link provided in github-onboarding. The install link format is `https://github.com/apps/{appSlug}/installations/new`. Include this as a clickable link.
   - **Self-hosted:** Direct them to the Apps page (`/apps`) in their Nixopus dashboard. The setup wizard there will create a GitHub App automatically via the manifest flow.
4. Tell them to come back after connecting so you can list their repos and deploy.

## Rules

- Do NOT call any other tools in this response — only `read_skill("github-onboarding")` to get accurate steps.
- Do NOT invent navigation paths like "Settings → Integrations → GitHub" — this does not exist.
- Do NOT ask questions in the first response — just greet and give the next step.
- Keep the total response to 4-6 sentences.
- The [user-context] block ALWAYS contains `instance: mode=cloud` or `instance: mode=self-hosted`. Use it directly. NEVER ask the user which one they are on.
- If mode is "cloud": give the GitHub App install link.
- If mode is "self-hosted": direct them to the Apps page (`/apps`).

## After the user returns

Once the user says they've connected GitHub (or comes back), call `get_github_connectors` to verify, then proceed with listing repositories and guiding the first deployment.

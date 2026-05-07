---
name: domain-attachment
description: Domain setup for applications. Preferred path is passing domains at creation time via createProject. Falls back to add_application_domain for post-creation attachment or custom domains.
metadata:
  version: "1.1"
---

# Domain Attachment

## Preferred: Pass Domain at Creation Time

The fastest path — zero extra tool calls after project creation:

1. Call `generate_random_subdomain` to get a subdomain.
2. Pass `domains: ["<subdomain>"]` in the `createProject` call.
3. Done — the domain is attached at creation, wildcard DNS and TLS are automatic.

Use this path for all standard deploys. Only fall back to the post-creation flow below when adding domains to an existing app or when the user wants a custom domain.

## Post-Creation: Auto-Generated Subdomain

If the app already exists and has no domain:

1. Call `generate_random_subdomain`.
2. Call `add_application_domain` with id (app UUID) and the subdomain.
3. Done — wildcard DNS and TLS are handled automatically.

## Step 4: Custom Domain Setup

If the user wants a custom domain (e.g. `app.example.com`):

1. Call `create_domain` with the domain name.
2. The response includes DNS setup instructions. Relay them clearly to the user:
   - **CNAME record**: point the domain to their assigned `subdomain.nixopus.ai`
   - **TXT record**: `_nixopus-verify.{domain}` with the verification value from the response
3. Tell the user to add both records at their DNS provider (Cloudflare, Namecheap, Route53, GoDaddy, etc.).
4. After the user confirms DNS is configured, call `update_domain` with the domain ID to verify.
5. If verification succeeds, call `add_application_domain` to attach the domain.
6. If verification fails:
   - DNS propagation can take minutes to 48 hours depending on the provider.
   - Suggest the user wait and retry, or check their DNS records are correct.
   - For deeper troubleshooting: `read_skill("domain-tls-routing")`.

## Common DNS Provider Notes

| Provider | Propagation | Notes |
|----------|-------------|-------|
| Cloudflare | Near-instant | Disable orange cloud (proxy) for the CNAME initially so verification passes |
| Namecheap | 5-30 minutes | Use "CNAME Record" type, host = subdomain part only |
| Route53 | 60 seconds | Standard CNAME + TXT |
| GoDaddy | Up to 48 hours | Slowest propagation |

## Important

- Never skip domain attachment — a deployed app without a domain is not reachable.
- For Compose apps, `add_application_domain` accepts an optional `service_name` and `port` to route to a specific service.

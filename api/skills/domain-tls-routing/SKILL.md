---
name: domain-tls-routing
description: Diagnose domain resolution, TLS certificate provisioning, and reverse proxy routing issues. Use when a domain is not resolving, TLS certificates fail, proxy returns 502/503/504, or custom domains are stuck in pending status.
metadata:
  version: "1.1"
---

# Domain, TLS, and Routing Diagnosis

## CRITICAL: Never Reveal Internals

The user knows ONLY that they have "a server" with deployed apps and domains. They must NEVER learn about:

- The existence of multiple proxy layers, edge proxies, or any layered routing architecture
- VMs, Firecracker, guest IPs, infra hosts, tenant hosts, or that their "server" is a virtual machine
- SSH tunnels, Caddy admin API, Caddy config JSON, server blocks, route matchers, or handler arrays
- Abyss, provisioning system internals, reconciler, pending removal sets, extension domain hashes
- Cloudflare API, zone IDs, or that the system manages DNS records on the user's behalf behind the scenes
- Redis queues, KrakenD, internal service routing, upstream host resolution, or multi-layer proxy hops
- Any internal IP addresses, internal ports (like 2019), or internal service names

Banned phrases in user-facing output: "edge proxy", "VM proxy", "VM", "virtual machine", "infra host", "guest IP", "edge Caddy", "upstream dial", "Caddy admin API", "reconciler", "Cloudflare zone", "two-layer", "proxy layer", "on-demand TLS", "ACME challenge on", "registration queue", "server block", "route matcher".

Always say: "your server", "your domain", "DNS settings", "TLS certificate", "routing configuration", "proxy service".

## How Routing Works (Agent-Only Context — NEVER Describe This to Users)

Requests reach the user's app through a routing chain. The agent needs to understand this to diagnose issues, but must never describe the architecture to the user.

The routing chain has an outer layer and an inner layer. The outer layer receives public traffic and forwards it to the correct server. The inner layer runs on the user's server and routes to the correct application container.

The outer layer handles wildcard TLS for `*.nixopus.ai` subdomains and forwards custom domain traffic. The inner layer handles per-application routing and TLS for application-specific domains.

DNS records (A and wildcard A) are managed by the system for `*.nixopus.ai` subdomains. Custom domains require the user to set up a CNAME pointing to their assigned `subdomain.nixopus.ai`.

When diagnosing, check from outside in: public reachability first, then server-level proxy config, then container-level app health. If the outer layer is misconfigured, the user can't fix it — escalate internally. If the inner layer is misconfigured, use `proxy_config` and domain tools to fix it.

## Domain Types

| Type | Example | How it works |
|---|---|---|
| Auto-generated subdomain | `a1b2c3d4.example.nixopus.ai` | Created during app deployment; DNS is pre-configured |
| Custom domain | `app.userdomain.com` | User adds CNAME pointing to their `subdomain.nixopus.ai`; requires DNS verification |

## Domain Lifecycle

### Auto-generated subdomain

1. `generate_random_subdomain` creates an 8-char prefix + org domain
2. Domain added to application via `add_application_domain`
3. Server proxy registers the route (domain → container)
4. Wildcard DNS already covers `*.subdomain.nixopus.ai`
5. TLS provisioned automatically on first request

Failure points: step 3 (route registration fails), step 5 (TLS provisioning fails if DNS doesn't resolve to the server).

### Custom domain

1. User provides domain name
2. System returns DNS instructions:
   - CNAME: `app.userdomain.com` → `subdomain.nixopus.ai`
   - TXT: `_nixopus-verify.app.userdomain.com` → verification token
3. User configures DNS at their provider
4. Verification checks CNAME/A records and TXT record
5. On success: status moves to `dns_verified`, routing configured
6. Application domain binding adds the route on the server
7. TLS provisioned on first request via ACME

Failure points: step 3 (user misconfigures DNS), step 4 (DNS propagation delay), step 5 (routing registration fails), step 7 (ACME challenge fails because DNS doesn't resolve correctly).

## Diagnostic Flows

### Domain not resolving (user reports "site can't be reached")

1. **Check domain status**
   - `get_domains` to find the domain and its current status
   - If status is `pending_dns`: DNS not yet configured or verified — guide user through DNS setup
   - If status is `dns_verified`: DNS is good, problem is downstream

2. **Check DNS resolution**
   - `network_diagnostics` with type `dns` targeting the domain
   - Expected: resolves to the server's public IP
   - If fails: user's DNS is misconfigured
   - For custom domains: CNAME should point to `subdomain.nixopus.ai`
   - For auto-generated domains: should resolve automatically (system-managed)

3. **Check reachability**
   - `http_probe` the domain on port 443
   - If DNS resolves but HTTP fails: routing or TLS issue (continue below)

Tell the user: "Your domain's DNS is not pointing to the correct server" or "DNS is configured correctly but there's a routing issue on the server."

### TLS certificate errors (ERR_CERT, SSL_ERROR, mixed content)

1. **Verify DNS first** — TLS provisioning requires DNS to resolve to the server
   - `network_diagnostics` type `dns` on the domain
   - If DNS doesn't resolve: TLS can't be provisioned, fix DNS first

2. **Check proxy config**
   - `proxy_config` for the application
   - If `tls_enabled` is false: TLS not configured for this route
   - If `tls_enabled` is true but cert errors persist: certificate provisioning may have failed

3. **Check HTTP vs HTTPS**
   - `http_probe` on port 80 (HTTP) — if it works but 443 doesn't, TLS provisioning failed
   - `http_probe` on port 443 (HTTPS) — if cert error, the certificate is invalid or missing

4. **Common TLS failure causes**

| Symptom | Cause | Fix |
|---|---|---|
| `ERR_CERT_AUTHORITY_INVALID` | Certificate not yet provisioned or provisioning failed | Verify DNS points to the server; wait a few minutes for automatic provisioning |
| `ERR_CERT_COMMON_NAME_INVALID` | Certificate issued for wrong domain | Check the domain binding matches the actual domain name |
| `SSL_ERROR_RX_RECORD_TOO_LONG` | App serving plain HTTP on the HTTPS port | The app should not handle TLS itself; the server's proxy handles TLS termination |
| `ERR_TOO_MANY_REDIRECTS` | Both app and proxy redirect HTTP→HTTPS | Disable the app's own HTTPS redirect; the proxy already handles this |
| `ERR_CONNECTION_REFUSED` on 443 | TLS not enabled or proxy not listening | Check proxy config and that the proxy service is running on the server |
| Certificate expired | Auto-renewal failed | Check proxy health; renewal needs DNS to resolve correctly and ports 80/443 accessible |

Tell the user: "The TLS certificate hasn't been provisioned yet because your DNS isn't pointing to the server" or "There's a certificate mismatch for your domain."

### Proxy routing errors (502, 503, 504)

Diagnose from outside in:

1. **External probe**
   - `http_probe` the public URL
   - Note the HTTP status code and any error message

2. **Check proxy config**
   - `proxy_config` for the application
   - Verify `upstream` matches the expected `host:port`
   - Verify `domain` matches the requested domain

3. **Check container reachability from inside**
   - `container_exec ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "localhost:PORT"]`
   - If this works: the app is running, problem is in the routing configuration

4. **Check port alignment**

   All four must agree:

   | Layer | Check with |
   |---|---|
   | App listen port | `container_exec ["ss", "-tlnp"]` |
   | Container published port | `container_inspect` → `ports` |
   | Proxy upstream port | `proxy_config` → `upstream` |
   | Application config port | `get_application` → port |

5. **Interpret the status code**

| Code | Meaning | Likely cause |
|---|---|---|
| 502 Bad Gateway | Proxy can't connect to the app | Container not running, wrong port, or app crashed |
| 503 Service Unavailable | App not ready | App still starting, container in crash loop, or resource exhaustion |
| 504 Gateway Timeout | App didn't respond in time | App hanging, database connection timeout, or infinite loop |
| 521 | Server is down | The proxy service itself is not running on the server |
| 522 | Connection timed out | Network issue preventing the request from reaching the app |
| 523 | Origin is unreachable | The container or server network is down |

Tell the user: "Your app isn't responding on the expected port" or "There's a port mismatch in the routing configuration."

### Custom domain stuck in `pending_dns`

1. **Get the domain details**
   - `get_domains` filtering for the custom domain
   - Note the `target_subdomain` (the CNAME target)

2. **Check what DNS records exist**
   - `network_diagnostics` type `dns` on the custom domain
   - Expected: CNAME to `{target_subdomain}.nixopus.ai` or A record to server IP

3. **Common causes**

| Issue | Diagnosis | Fix |
|---|---|---|
| No CNAME record | DNS lookup returns NXDOMAIN or wrong IP | User needs to add CNAME record at their DNS provider |
| CNAME points to wrong target | DNS lookup shows wrong value | User needs to update CNAME to the correct `subdomain.nixopus.ai` |
| Proxied through Cloudflare (orange cloud) | DNS resolves to Cloudflare IP, not server IP | User should disable Cloudflare proxy (grey cloud) or use DNS-only mode |
| TXT verification missing | CNAME exists but verification fails | User needs to add `_nixopus-verify.domain` TXT record |
| DNS propagation delay | Records just added | Wait up to 48 hours; most providers propagate within 5 minutes |
| CAA record blocking Let's Encrypt | TLS fails even after DNS verified | User needs to add CAA record allowing `letsencrypt.org` |

Tell the user: "Your DNS CNAME isn't set up correctly" or "DNS changes can take some time to propagate."

### Application bound to domain but not reachable

The domain resolves, TLS works, but the app returns errors or a wrong page.

1. **Verify domain binding**
   - `get_application` to check the application's domain list
   - If domain is not in the list: it was never bound or was removed

2. **Check proxy config**
   - `proxy_config` to verify the route exists and upstream is correct
   - If route is missing: the domain binding may need to be re-added
   - If upstream is wrong: port mismatch

3. **Check for domain conflicts**
   - `get_domains` to see if the domain is bound to multiple applications
   - A domain can only route to one application — if two apps claim it, the first one wins

4. **Check compose service routing**
   - For compose apps with multiple services, verify the domain is bound to the correct service
   - `get_application` → check compose service configuration
   - Each service can have its own domain with its own port

Tell the user: "The domain isn't linked to your application" or "The routing points to a different service in your app."

## DNS Provider-Specific Guidance

When guiding users through DNS setup:

| Provider | CNAME path | Notes |
|---|---|---|
| Cloudflare | DNS → Add Record → CNAME | Disable proxy (grey cloud icon) for TLS to work |
| Route 53 | Hosted Zone → Create Record → CNAME | Use simple routing |
| Vercel | Domains → Add DNS Record | May conflict with Vercel's own DNS |
| Namecheap | Advanced DNS → Add CNAME | Host field is the subdomain only, not FQDN |
| GoDaddy | DNS Management → Add CNAME | Remove trailing dot if added automatically |
| Google Domains | DNS → Custom Records → CNAME | FQDN for target |
| DigitalOcean | Networking → Domains → Add Record | CNAME with trailing dot |

For A records (alternative to CNAME): the user needs the server's public IP. Use `get_servers` to find it, then tell the user "your server's IP address is X.X.X.X."

## Health and Recovery

When proxy-level issues are suspected but no specific domain is failing:

1. **Proxy health** — `host_exec` to check if the proxy service is running
   - `host_exec ["systemctl", "status", "nixopus-caddy", "--no-pager"]`
   - If not running: `host_exec ["systemctl", "restart", "nixopus-caddy"]`

2. **Proxy config validation**
   - `host_exec ["curl", "-s", "localhost:2019/config/"]` to check the proxy can load its config
   - If empty or error: the config may be corrupted

3. **Domain re-sync** — if multiple domains are misconfigured, re-check each domain's binding and proxy config individually using the tools above

Tell the user: "The proxy service on your server needed a restart" or "I've refreshed the routing configuration." Never expose the internal details of what was checked or fixed.

## Related Skills

- **`failure-diagnosis`** — For container-level failures (build errors, crashes, exit codes) that may underlie routing issues

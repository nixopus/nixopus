---
name: caddyfile-generation
description: Generate Caddyfile configurations for static sites and reverse proxies — SPA fallback routing, cache headers, compression, redirects, and error pages. Use when deploying a static site that needs custom Caddy configuration, or when the user needs SPA routing, caching, or redirect rules.
metadata:
  version: "1.0"
---

# Caddyfile Generation

## When to Generate

- Static sites served by Caddy (from `static-deploy` or frontend builds)
- SPA applications that need fallback routing (React, Vue, Angular, SvelteKit static)
- Sites needing custom cache headers, compression, or redirects

## Base Template

Minimal Caddyfile for a static site:

```caddyfile
:80 {
	root * /srv
	file_server
}
```

## SPA Fallback Routing

Single-page apps need all non-file routes to return `index.html`:

```caddyfile
:80 {
	root * /srv
	try_files {path} /index.html
	file_server
}
```

Frameworks that need this: React (CRA, Vite), Vue, Angular, SvelteKit (static adapter), Remix (SPA mode).

Frameworks that do NOT need this: Next.js (has its own server), Nuxt (SSR), Astro (generates individual HTML files for each route).

## Cache Headers for Hashed Assets

Frontend build tools (Vite, Webpack) produce files with content hashes (e.g. `main.a1b2c3.js`). These can be cached aggressively:

```caddyfile
:80 {
	root * /srv

	@hashed path_regexp hashed \.(js|css|woff2?|ttf|eot|svg|png|jpg|jpeg|gif|ico|webp)$
	header @hashed Cache-Control "public, max-age=31536000, immutable"

	@html path *.html /
	header @html Cache-Control "no-cache, no-store, must-revalidate"

	try_files {path} /index.html
	file_server
}
```

HTML files must NOT be cached (they reference the hashed assets).

## Compression

Enable gzip and zstd compression:

```caddyfile
:80 {
	root * /srv
	encode zstd gzip
	try_files {path} /index.html
	file_server
}
```

## Redirects

### www to non-www

```caddyfile
www.example.com {
	redir https://example.com{uri} permanent
}

example.com {
	root * /srv
	file_server
}
```

### HTTP to HTTPS

Caddy handles this automatically when using domain names. Only needed for explicit port-based configs:

```caddyfile
:80 {
	redir https://{host}{uri} permanent
}
```

Note: In Nixopus deployments, the proxy layer already handles TLS termination. Do NOT add HTTPS redirects in the app's Caddyfile — the proxy handles this.

## Custom Error Pages

```caddyfile
:80 {
	root * /srv

	handle_errors {
		@404 expression {err.status_code} == 404
		rewrite @404 /404.html
		file_server
	}

	try_files {path} /index.html
	file_server
}
```

## Reverse Proxy (API backend)

When a static frontend needs to proxy API requests to a backend:

```caddyfile
:80 {
	root * /srv

	handle /api/* {
		reverse_proxy backend:8080
	}

	try_files {path} /index.html
	file_server
}
```

## Security Headers

```caddyfile
:80 {
	root * /srv

	header {
		X-Content-Type-Options "nosniff"
		X-Frame-Options "DENY"
		Referrer-Policy "strict-origin-when-cross-origin"
		-Server
	}

	file_server
}
```

## Complete Production Template

Combines SPA routing, caching, compression, and security headers:

```caddyfile
:80 {
	root * /srv
	encode zstd gzip

	header {
		X-Content-Type-Options "nosniff"
		X-Frame-Options "DENY"
		Referrer-Policy "strict-origin-when-cross-origin"
		-Server
	}

	@hashed path_regexp hashed \.(js|css|woff2?|ttf|eot|svg|png|jpg|jpeg|gif|ico|webp)$
	header @hashed Cache-Control "public, max-age=31536000, immutable"

	@html path *.html /
	header @html Cache-Control "no-cache, no-store, must-revalidate"

	try_files {path} /index.html
	file_server
}
```

## Generation Logic

1. Determine if the app is a SPA (needs `try_files` fallback) or multi-page (doesn't)
2. Start with the base template
3. Add `try_files {path} /index.html` if SPA
4. Add `encode zstd gzip` for compression
5. Add hashed asset cache headers if the build tool produces hashed filenames
6. Add security headers
7. Write to `Caddyfile` at the project root

## Gotchas

- Caddy listens on `:80` inside Docker — the proxy layer handles external port mapping and TLS
- `try_files` must come BEFORE `file_server` in the Caddyfile
- `root * /srv` must match the COPY destination in the Dockerfile
- Do NOT configure HTTPS/TLS in the Caddyfile for Nixopus deployments — the edge proxy handles TLS termination
- Caddy's `file_server` serves directory listings by default — add `file_server browse` only if intentional

## Related Skills

- **`static-deploy`** — Dockerfile patterns that use Caddy as the web server
- **`dockerfile-generation`** — Generate the Caddyfile alongside the Dockerfile

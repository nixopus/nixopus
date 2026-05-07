---
name: static-deploy
description: Deploy static file sites — Caddy/nginx serving, Staticfile config, and Dockerfile patterns. Use when deploying a static HTML site with no server-side runtime, or when index.html or a Staticfile is detected at the project root.
metadata:
  version: "1.0"
---

# Static Files Deployment

## Detection

Project is static if:

1. `index.html` exists in the root
2. Or `Staticfile` exists at the project root (explicit configuration)

## Configuration

### Root Directory

1. `Staticfile` contents: `root: <dir>`
2. Default: current directory (`.`)

### Staticfile

Create a `Staticfile` at the project root:

```
root: dist
```

### Custom Caddyfile

Place a `Caddyfile` at the project root to override the default Caddy configuration.

## Typical Layouts

| Layout | Root to serve |
|---|---|
| Built SPA (Vite, etc.) | `dist` |
| Jekyll / Hugo output | `_site` or `public` |
| Plain HTML | `.` (current dir) |
| Nested | Configure via Staticfile |

## Dockerfile Patterns

### Nginx

```dockerfile
FROM nginx:alpine
COPY . /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Caddy

```dockerfile
FROM caddy:alpine
COPY . /srv
EXPOSE 80
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
```

### With custom Caddyfile

```dockerfile
FROM caddy:alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY dist /srv
EXPOSE 80
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
```

---
name: env-detection
description: Detect required environment variables from source code, config files, and .env examples. Use when preparing for deployment, checking for missing env vars, or when the user asks about required environment configuration.
metadata:
  version: "1.0"
---

# Environment Variable Detection

Identify all environment variables an application needs before deployment. Missing env vars are the most common cause of deployment failures.

## Detection sources (check in order)

### 1. `.env.example` / `.env.sample` / `.env.template`

Primary source. Read the file and extract every variable name and any inline comments describing its purpose.

```
DATABASE_URL=postgresql://user:pass@localhost:5432/db
REDIS_URL=redis://localhost:6379
API_KEY=your-api-key-here
SECRET_KEY=change-me
```

### 2. Source code patterns

Use `grep` across the repo root for these patterns:

| Pattern | Language |
|---------|----------|
| `process.env.VAR_NAME` | Node.js |
| `Deno.env.get("VAR_NAME")` | Deno |
| `os.environ["VAR_NAME"]` or `os.getenv("VAR_NAME")` | Python |
| `os.Getenv("VAR_NAME")` | Go |
| `ENV["VAR_NAME"]` or `ENV.fetch("VAR_NAME")` | Ruby |
| `env::var("VAR_NAME")` | Rust |
| `System.getenv("VAR_NAME")` | Java |
| `@Value("${VAR_NAME}")` | Spring |

### 3. Framework config files

| File | What to look for |
|------|-----------------|
| `next.config.js` / `next.config.mjs` | `env:` block, `NEXT_PUBLIC_*` prefixed vars |
| `nuxt.config.ts` | `runtimeConfig` block |
| `vite.config.ts` | `define` block, `VITE_*` prefixed vars |
| `docker-compose.yml` | `environment:` sections |
| `Dockerfile` | `ENV` and `ARG` directives |
| `settings.py` (Django) | `os.environ` calls |
| `config/*.yml` (Rails) | `<%= ENV["VAR"] %>` patterns |

### 4. Database/service URLs

Check dependency manifests for services that typically require connection URLs:

| Dependency | Expected env var |
|-----------|-----------------|
| `pg` / `sequelize` / `prisma` / `typeorm` | `DATABASE_URL` |
| `mongoose` / `mongodb` | `MONGODB_URI` |
| `ioredis` / `redis` | `REDIS_URL` |
| `@aws-sdk/*` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| `stripe` | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` |
| `@sendgrid/mail` | `SENDGRID_API_KEY` |
| `nodemailer` | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS` |
| `@auth0/*` / `next-auth` | `AUTH_SECRET`, `AUTH_URL`, provider-specific keys |

## Classification

Classify each detected variable:

| Category | Examples | Required for deploy? |
|----------|---------|---------------------|
| **Infrastructure** | `DATABASE_URL`, `REDIS_URL`, `PORT` | Yes |
| **Secrets** | `API_KEY`, `SECRET_KEY`, `JWT_SECRET` | Yes |
| **Service URLs** | `NEXT_PUBLIC_API_URL`, `WEBHOOK_URL` | Yes (but values differ per environment) |
| **Build-time** | `NEXT_PUBLIC_*`, `VITE_*` | Yes (must be set during build) |
| **Optional/Debug** | `LOG_LEVEL`, `DEBUG`, `NODE_ENV` | No (has sensible defaults) |

## Build-time vs runtime

This distinction matters for Dockerfiles:

- **Build-time**: Set as `ARG` in Dockerfile, passed via `--build-arg` or `args:` in compose. Includes `NEXT_PUBLIC_*`, `VITE_*`, and any var used during `npm run build`.
- **Runtime**: Set as `ENV` in Dockerfile or `environment:` in compose. Includes `DATABASE_URL`, `PORT`, API keys.

## Output format

After detection, report:

1. List of all detected env vars with their source (`.env.example`, source code, framework config)
2. Classification (infrastructure, secret, service URL, build-time, optional)
3. Which vars are missing values (need user input)
4. Which vars have safe defaults (e.g. `PORT=3000`, `NODE_ENV=production`)
5. Which vars are build-time and must be in Dockerfile `ARG` directives

## Related Skills

- **`pre-deploy-checklist`** — Uses env detection results to validate deployment readiness
- **`dockerfile-generation`** — Build-time env vars identified here need `ARG` directives in the Dockerfile

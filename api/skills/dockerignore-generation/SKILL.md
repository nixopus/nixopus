---
name: dockerignore-generation
description: Generate ecosystem-specific .dockerignore files to reduce build context size and prevent secret leaks. Use when no .dockerignore exists, when the build context is large, or when secrets may be leaking into images.
metadata:
  version: "1.0"
---

# .dockerignore Generation

## Why It Matters

Without a `.dockerignore`:
- Build context includes `node_modules` (hundreds of MB), `.git` history, and local env files
- Secrets in `.env` files get copied into the image and are extractable
- Build is slow because Docker sends the entire directory to the daemon

## Base Template (all ecosystems)

Every `.dockerignore` should include:

```
.git
.gitignore
.env
.env.*
!.env.example
!.env.sample
*.md
!README.md
LICENSE
docker-compose*.yml
.dockerignore
Dockerfile
.vscode
.idea
.cursor
```

## Ecosystem-Specific Entries

### Node.js

```
node_modules
.next
.nuxt
.output
dist
build
.cache
coverage
.nyc_output
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.pnpm-debug.log*
.turbo
.vercel
.netlify
storybook-static
```

### Python

```
__pycache__
*.pyc
*.pyo
*.egg-info
.eggs
.venv
venv
env
.tox
.pytest_cache
.mypy_cache
.ruff_cache
htmlcov
*.cover
```

### Go

```
vendor/
*.test
*.out
bin/
tmp/
```

### Rust

```
target/
*.rs.bk
```

### Java

```
target/
build/
.gradle/
*.class
*.jar
!*.jar  # if copying JARs intentionally, remove this line
.settings/
.classpath
.project
```

### Ruby

```
vendor/bundle
.bundle
log/
tmp/
coverage/
spec/reports
```

### PHP

```
vendor/
storage/logs/
storage/framework/cache/
storage/framework/sessions/
storage/framework/views/
bootstrap/cache/
```

### Elixir

```
_build/
deps/
.elixir_ls/
cover/
```

### .NET

```
bin/
obj/
*.user
*.suo
packages/
```

## Generation Logic

1. Start with the base template
2. Detect ecosystem from the project (check for `package.json`, `go.mod`, `requirements.txt`, etc.)
3. Append the matching ecosystem entries
4. If `test/` or `tests/` or `__tests__/` exists: add test directories
5. If `.github/` exists: add `.github/`
6. Write to `.dockerignore` at the project root

## Gotchas

- `!.env.example` negates the `.env.*` exclusion — keep example env files so Dockerfile can reference them
- Monorepos: `.dockerignore` is relative to the build context root, not the Dockerfile location
- Docker Compose `build.context` changes what `.dockerignore` applies to — if context is `.`, the root `.dockerignore` applies
- Don't ignore `prisma/` if Prisma is used — `prisma/schema.prisma` is needed for `postinstall`
- Don't ignore lockfiles (`package-lock.json`, `yarn.lock`, etc.) — they're essential for reproducible builds

## Related Skills

- **`pre-deploy-checklist`** — Checks for `.dockerignore` existence and flags missing ones
- **`dockerfile-generation`** — Generate `.dockerignore` alongside the Dockerfile

---
name: deployment-analysis
description: Analyze a repository to determine ecosystem, deployment targets, ports, build commands, and monorepo structure. Use when starting a new deployment, onboarding a repository, or when the user asks what stack or framework a project uses.
metadata:
  version: "1.0"
---

# Deployment Analysis

When analyzing a repository for deployment, follow this sequence using workspace tools.

## Step 1: Project structure

Run `list_directory` on the repo root. Look for:

| File | Meaning |
|------|---------|
| `package.json` | Node.js (check `engines`, `scripts.start`, `scripts.build`) |
| `go.mod` | Go |
| `requirements.txt` / `pyproject.toml` / `Pipfile` | Python |
| `Cargo.toml` | Rust |
| `pom.xml` / `build.gradle` | Java |
| `Gemfile` | Ruby |
| `mix.exs` | Elixir |
| `Dockerfile` | Already containerized |
| `docker-compose.yml` / `docker-compose.yaml` | Multi-service |
| `.env.example` / `.env.sample` | Env vars documented |

## Step 2: Detect ecosystem and framework

Read the manifest file to identify the framework:

**Node.js** — read `package.json`:
- `dependencies.next` → Next.js. Default port 3000. Build: `npm run build`. Start: `npm start`.
- `dependencies.nuxt` → Nuxt. Default port 3000.
- `dependencies.react-scripts` → Create React App. Static build. Port 80 (nginx).
- `dependencies.vite` → Vite app. Build outputs to `dist/`. Static or SSR depending on config.
- `dependencies.express` or `dependencies.fastify` or `dependencies.hono` → API server. Check `scripts.start` for port.
- `dependencies.@remix-run/node` → Remix. Port 3000.
- `dependencies.astro` → Astro. Check if SSR or static.

**Go** — read `go.mod`: Check module path. Look for `main.go` or `cmd/` directory. Default port 8080.

**Python** — read `requirements.txt` or `pyproject.toml`:
- `django` → Django. Default port 8000. Start: `gunicorn` or `python manage.py runserver`.
- `flask` → Flask. Default port 5000. Start: `gunicorn app:app`.
- `fastapi` → FastAPI. Default port 8000. Start: `uvicorn main:app`.

**Rust** — read `Cargo.toml`: Check `[dependencies]` for `actix-web`, `axum`, `rocket`. Default port 8080.

## Step 3: Detect port

Priority order for port detection:
1. `Dockerfile` — look for `EXPOSE` directive
2. `docker-compose.yml` — look for `ports:` mapping
3. `.env.example` — look for `PORT=`
4. Manifest file — check start script for `--port`, `-p`, or `PORT` references
5. Source code — `grep("listen|EXPOSE|PORT", repoRoot)` for hardcoded ports
6. Framework default (see table above)

## Step 4: Monorepo detection

Signs of a monorepo:
- `apps/` or `packages/` or `services/` directories at root
- `turbo.json` or `nx.json` or `lerna.json` or `pnpm-workspace.yaml`
- Multiple `package.json` files in subdirectories
- Multiple `Dockerfile` files in subdirectories

For monorepos:
- Each service under `apps/` or `services/` is a separate deployment target
- Check each service's manifest for its own port and build command
- `docker-compose.yml` is likely needed
- Check for shared dependencies in root `package.json`

## Step 5: Build command detection

| Ecosystem | Install | Build | Start |
|-----------|---------|-------|-------|
| Node (npm) | `npm install` | `npm run build` | `npm start` |
| Node (yarn) | `yarn install` | `yarn build` | `yarn start` |
| Node (pnpm) | `pnpm install` | `pnpm build` | `pnpm start` |
| Go | `go mod download` | `go build -o app ./...` | `./app` |
| Python (pip) | `pip install -r requirements.txt` | n/a | `gunicorn`/`uvicorn` |
| Python (poetry) | `poetry install` | n/a | `poetry run` |
| Rust | n/a | `cargo build --release` | `./target/release/<name>` |
| Java (Maven) | `mvn install` | `mvn package` | `java -jar target/*.jar` |
| Java (Gradle) | `gradle build` | `gradle build` | `java -jar build/libs/*.jar` |

Detect package manager: check for `yarn.lock` (yarn), `pnpm-lock.yaml` (pnpm), `package-lock.json` (npm), `bun.lockb` (bun).

## Output

After analysis, you should know:
- Ecosystem and framework
- Port number
- Build and start commands
- Whether a Dockerfile exists
- Whether docker-compose is needed
- What env vars are required

## Related Skills

- **`env-detection`** — After analysis, detect required environment variables
- **`dockerfile-generation`** — Generate a Dockerfile based on the detected ecosystem
- **`pre-deploy-checklist`** — Validate deployment readiness before triggering a build
- Language-specific skills (`node-deploy`, `python-deploy`, `go-deploy`, etc.) — Detailed build and deploy instructions for the detected ecosystem

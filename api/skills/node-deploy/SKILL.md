---
name: node-deploy
description: Build and deploy Node.js applications — version detection, package managers, framework-specific builds, monorepo support, and Dockerfile patterns. Use when deploying a Node.js, JavaScript, or TypeScript project, or when package.json is detected in the repository.
metadata:
  version: "1.0"
---

# Node.js Deployment

## Detection

Project is Node.js if `package.json` exists in the root directory.

## Versions

Node.js version priority:

1. `.node-version` or `.nvmrc` file
2. `engines.node` field in `package.json`
3. `mise.toml` or `.tool-versions`
4. Defaults to **22**

### Bun

If Bun is detected as the package manager:

1. `engines.bun` field in `package.json`
2. `.bun-version` file
3. `mise.toml` or `.tool-versions`
4. Defaults to **latest**

When Bun is the primary runtime, Node.js must still be installed if:

- The project uses Corepack (`packageManager` field exists in `package.json`)
- Build tools require Node (Astro, Vite, or native addon compilation via `node-gyp`)
- Any `package.json` script explicitly invokes `node`

## Package Managers

Detect in order:

1. **`packageManager` field** in `package.json` → use Corepack to install exact version (e.g. `pnpm@9.1.0`)
2. **Lock files:**

| Lock file | Package manager |
|---|---|
| `package-lock.json` | npm |
| `yarn.lock` | Yarn (check `.yarnrc.yml` to distinguish Yarn Berry from Classic) |
| `pnpm-lock.yaml` | pnpm |
| `bun.lockb` or `bun.lock` | Bun |

3. **`engines` field** — `engines.pnpm` → pnpm, `engines.bun` → Bun, `engines.yarn` → Yarn
4. **Default:** npm

### Install Commands

| Package manager | With lockfile | Without lockfile |
|---|---|---|
| npm | `npm ci` | `npm install` |
| yarn (classic) | `yarn install --frozen-lockfile` | `yarn install` |
| yarn (berry) | `yarn install --immutable` | `yarn install` |
| pnpm | `pnpm install --frozen-lockfile` | `pnpm install` |
| bun | `bun install --frozen-lockfile` | `bun install` |

## Runtime Variables

Set `NODE_ENV=production` for the runtime stage. During the build, keep `NPM_CONFIG_PRODUCTION=false` and `YARN_PRODUCTION=false` so dev dependencies remain available for compilation. Disable update notifications with `NPM_CONFIG_UPDATE_NOTIFIER=false` and `NPM_CONFIG_FUND=false`. Set `CI=true` to enable CI-appropriate behavior in tooling.

## Build & Start

### Start Command Resolution

1. `start` script in `package.json`
2. `main` or `module` field in `package.json` (run with `node`)
3. `server.js`, `index.js`, or `index.ts` in root (run with `node`)

### Build Command Resolution

1. `build` script in `package.json` → `${packageManager} run build`
2. If no build script → skip build step

### Output Directory

| Framework | Output directory |
|---|---|
| NestJS | `dist` |
| Next.js (SSR) | `.next` |
| Next.js (export) | `out` |
| Nuxt | `.output` |
| SvelteKit | `build` |
| Remix | `build` |
| Astro | `dist` |
| Vite | `dist` |
| Angular | `dist/${projectName}` |
| React (CRA) | `build` |
| React Router | `build/client` |
| Default | `dist` |

## Port Detection

1. **Environment files** — `PORT=<number>` from `.env`, `.env.example`, `.env.production`
2. **`package.json` scripts** — scan `start`, `dev`, `serve` for `-p <port>`, `--port <port>`, `PORT=<port>`
3. **Framework config** — `next.config.*` or `vite.config.*` for `port: <number>`
4. **Framework defaults:**

| Framework | Default port |
|---|---|
| Express | 3000 |
| Fastify | 3000 |
| NestJS | 3000 |
| Hono | 3000 |
| Next.js | 3000 |
| Nuxt | 3000 |
| Remix | 3000 |
| SvelteKit | 5173 |
| Astro | 4321 |
| Vite | 5173 |
| React | 3000 |
| Vue | 8080 |

5. **Final default:** 3000

## Framework Detection

From `package.json` dependencies (merge `dependencies` + `devDependencies`). First match wins.

| Package pattern | Framework | Category |
|---|---|---|
| `express` | Express | Backend |
| `fastify` | Fastify | Backend |
| `@nestjs/core` | NestJS | Backend |
| `hono` | Hono | Backend |
| `next` | Next.js | FullStack |
| `nuxt` | Nuxt | FullStack |
| `@sveltejs/kit` | SvelteKit | FullStack |
| `@remix-run/node` or `@remix-run/react` | Remix | FullStack |
| `astro` | Astro | Static |
| `vite` (without a higher framework) | Vite | Frontend |
| `react` + `react-dom` (without Next/Remix) | React | Frontend |
| `vue` (without Nuxt) | Vue | Frontend |

Config file fallback:

| Config file | Framework |
|---|---|
| `nest-cli.json` | NestJS |
| `next.config.js`, `next.config.mjs`, `next.config.ts` | Next.js |
| `nuxt.config.js`, `nuxt.config.ts` | Nuxt |
| `svelte.config.js` | SvelteKit |
| `remix.config.js` | Remix |
| `astro.config.mjs`, `astro.config.js` | Astro |
| `vite.config.ts`, `vite.config.js` | Vite |
| `vue.config.js` | Vue |
| `angular.json` | Angular |

### Framework-Specific Behavior

**Next.js**
- Check `next.config.*` for `output: "standalone"` → standalone build (smaller image, includes `node_modules` subset)
- `output: "export"` → static site, no server needed
- Cache `.next/cache` between builds
- `app/` directory → React Server Components

**Nuxt**
- Default start: `node .output/server/index.mjs`
- Cache `node_modules/.cache`

**Astro**
- If `output` is not `"server"` → static site
- Cache `node_modules/.astro`

## Monorepo Support

### Detection Signals

- `workspaces` field in root `package.json`
- `pnpm-workspace.yaml`
- Build orchestrators: `turbo.json`, `nx.json`, `lerna.json`, `rush.json`
- Conventional directories: `apps/`, `packages/`, `services/`

### Workspace Package Resolution

- **pnpm:** Parse `pnpm-workspace.yaml` → `packages:` list
- **npm / yarn / bun:** Parse `workspaces` field in root `package.json`

### Build Steps

1. Detect workspace configurations automatically
2. Install all workspace dependencies (copy all `package.json` files + root lock file)
3. Respect workspace dependency links
4. Cache workspace `node_modules`
5. Build the target workspace package

## Optimizing the Install Layer

**Always copy:**
- `package.json` (root + workspace packages if monorepo)
- Lock file
- `pnpm-workspace.yaml` (if pnpm monorepo)
- `.npmrc` (if exists — contains registry config)

**Framework-specific install files** (copy if they exist, they trigger postinstall):
- `prisma/schema.prisma` — Prisma generates client on `postinstall`
- `.env` files needed at build time (e.g. Next.js `NEXT_PUBLIC_*`)

If `package.json` defines `preinstall` or `postinstall` scripts that depend on source files, copy the entire source before install to avoid broken hooks.

## Static Sites

| Framework | Detection | Default output dir |
|---|---|---|
| CRA | `react-scripts` in deps | `build` |
| Vite | `vite.config.js/ts` or build script contains `vite build` | `dist` |
| Angular | `angular.json` | `dist/${projectName}` |
| Astro | `astro.config.*` and output is not `"server"` | `dist` |
| Next.js (export) | `output: "export"` in config | `out` |
| React Router | `react-router.config.*` (`ssr: false` for SPA) | `build/client` |

Serve with Caddy/nginx. SPA fallback, cache headers for hashed assets, gzip/brotli.

## Environment Variable Semantics

### Build-Time vs Runtime

| Framework | Build-time prefix | Runtime access |
|---|---|---|
| Next.js | `NEXT_PUBLIC_*` | `process.env.*` (server only) |
| Nuxt | `NUXT_PUBLIC_*` | `process.env.*` via `useRuntimeConfig()` |
| Vite | `VITE_*` | not available at runtime (build-only) |
| SvelteKit | `PUBLIC_*` | `$env/static/public` (build-only) |
| Astro | `PUBLIC_*` | `import.meta.env.*` (build-only) |
| CRA | `REACT_APP_*` | not available at runtime (build-only) |

Build-time env vars must be available during Docker `build` step (via `ARG` + `ENV`).

## System Dependencies

| Package | Required system packages |
|---|---|
| Puppeteer | Chromium, xvfb, font libraries, Chrome system deps |
| Playwright | Chromium headless shell, system packages |
| `sharp` | `libvips` and build tools |
| `bcrypt` | `python3`, `make`, `g++` |
| `canvas` | `libcairo2-dev`, `libjpeg-dev`, `libpango1.0-dev`, `libgif-dev`, `build-essential` |

## Dev Dependency Pruning

After build, remove dev dependencies to reduce image size:

- **npm:** `npm prune --omit=dev`
- **yarn:** `yarn install --production` or set `NODE_ENV=production` during install
- **pnpm:** `pnpm prune --prod`
- **bun:** `bun install --production`

Skip pruning if the start command references a dev dependency (`ts-node`, `tsx`, `nodemon`).

## Caching

| Framework | Cache directory |
|---|---|
| NestJS | `node_modules/.cache` |
| Next.js | `.next/cache` |
| Nuxt | `node_modules/.cache` |
| SvelteKit | `node_modules/.cache` |
| Remix | `.cache` |
| React Router | `.react-router` |
| Astro | `node_modules/.astro` |
| Vite | `node_modules/.vite` |
| Default | `node_modules/.cache` |

## Dockerfile Patterns

### Simple Node.js Server (Express, Fastify, NestJS, Hono)

```dockerfile
FROM node:<version>-slim AS base

FROM base AS deps
WORKDIR /app
COPY package.json <lockfile> ./
RUN <install-command>

FROM base AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN <build-command>

FROM base AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./
EXPOSE <port>
CMD ["node", "dist/index.js"]
```

### Next.js Standalone

```dockerfile
FROM node:<version>-slim AS base

FROM base AS deps
WORKDIR /app
COPY package.json <lockfile> ./
RUN <install-command>

FROM base AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN <build-command>

FROM base AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

### Static Site (Vite, CRA, Astro static)

```dockerfile
FROM node:<version>-slim AS build
WORKDIR /app
COPY package.json <lockfile> ./
RUN <install-command>
COPY . .
RUN <build-command>

FROM caddy:alpine AS runtime
COPY --from=build /app/<output-dir> /srv
COPY Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
```

## Gotchas

- Yarn Berry (v2+) uses Plug'n'Play by default — `node_modules` won't exist unless `nodeLinker: node-modules` is set in `.yarnrc.yml`
- `npm ci` deletes `node_modules` before installing — copy `package.json` + lockfile first, install, then copy source for proper layer caching
- Next.js `output: "standalone"` must be set in config BEFORE running the build — the build step generates the standalone directory
- Prisma runs `prisma generate` on `postinstall` — `prisma/schema.prisma` must be copied before `npm install`
- pnpm with `--shamefully-hoist` may be needed for packages expecting a flat `node_modules` layout
- Bun `--frozen-lockfile` is the correct flag (not `--ci` like npm)
- SvelteKit and Astro dev servers use different ports (5173, 4321) than production — ensure EXPOSE matches the production port

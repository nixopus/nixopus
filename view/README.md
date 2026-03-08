# Nixopus View

The frontend for Nixopus, built with Next.js 15 and React 19.

## Prerequisites

- Node.js 18+
- A running Nixopus API instance (see `../api/`)
- A running Nixopus Auth instance (or use `docker-compose-dev.yml`)

## Setup

1. **Install dependencies**

```bash
npm install
```

2. **Configure environment**

```bash
cp .env.sample .env
```

Edit `.env` with your values:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port for `next start` | `7443` |
| `NEXT_PUBLIC_PORT` | Port exposed via `/api/config` | `7443` |
| `API_URL` | Nixopus API base URL | `http://localhost:8080/api` |
| `AUTH_SERVICE_URL` | Better Auth service URL (server-side) | `http://localhost:9090` |
| `PASSWORD_LOGIN_ENABLED` | Show password login (`true`) or OTP login (`false`) | `true` |

Optional variables:

| Variable | Description |
|----------|-------------|
| `AGENT_URL` | AI agent service URL (enables AI chat features) |
| `BASE_PATH` | Next.js base path (e.g. `/app`) |
| `ASSET_PREFIX` | CDN/asset prefix for static files |
| `NEXT_PUBLIC_APP_URL` | Public app URL for GitHub App manifest |
| `NEXT_PUBLIC_REDIRECT_URL` | GitHub App OAuth callback URL |

WebSocket and webhook URLs are automatically derived from `API_URL`. To override, set `WEBSOCKET_URL` or `WEBHOOK_URL` explicitly.

3. **Start development server**

```bash
npm run dev
```

The app will be available at `http://localhost:3000`.

4. **Start production server**

```bash
npm run build
npm run start
```

The app will be available at `http://localhost:7443` (or your configured `PORT`).

## Development with Docker

The easiest way to run all dependencies (database, Redis, auth, Caddy) is with the dev compose file from the project root:

```bash
docker compose -f docker-compose-dev.yml up
```

This starts PostgreSQL, Redis, the auth service, and Caddy. Run the API and view locally for hot reloading.

## Available Scripts

```bash
npm run dev       # Start dev server (Turbopack)
npm run build     # Production build
npm run start     # Start production server
npm run lint      # Run ESLint
npm run format    # Format code with Prettier
npm run analyze   # Build with bundle analyzer
```

## Git Hooks

The repo uses Husky for pre-commit and commit-msg hooks:

- **Pre-commit**: Runs `go fmt` + `goimports` on staged `.go` files, and Prettier on staged view files (`.ts`, `.tsx`, `.js`, `.css`, `.md`, `.json`). Formatted files are automatically re-staged.
- **Commit message**: Validates commit messages with [commitlint](https://commitlint.js.org/) (conventional commits format).

Hooks are installed automatically via `npm install` (`prepare` script runs `husky`).

## Project Structure

```
view/
├── app/                    # Next.js app router
│   ├── api/                # API routes (auth proxy, config, agent proxy)
│   ├── auth/               # Login/register pages
│   └── apps/               # Application pages (dashboard, deploy, etc.)
├── packages/
│   ├── components/         # Reusable UI components
│   ├── hooks/              # Custom React hooks
│   ├── lib/                # Utilities (auth client, agent client)
│   └── types/              # TypeScript type definitions
├── redux/                  # Redux store, slices, and API services
├── lib/i18n/               # Internationalization (en, es, fr, ml, kn)
└── public/                 # Static assets
```

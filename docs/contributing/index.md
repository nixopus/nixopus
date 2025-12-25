# Development Guide

Quick setup guide for Nixopus development.

## Prerequisites

Before you begin, ensure you have the following installed:

::: info Prerequisites Checklist
- **Go** 1.23.6 or higher
- **Node.js** 18 or higher  
- **Docker** and **Docker Compose** (handles PostgreSQL and SuperTokens)
:::

::: tip Verify Your Setup
You can verify your installations with:
```bash
go version
node --version
docker --version
docker compose version
```
:::

## Setup

### 1. Fork and Clone

First, fork the [nixopus repository](https://github.com/raghavyuva/nixopus) on GitHub, then clone your fork. Replace `your_username` with your actual GitHub username:

::: code-group

```bash [SSH]
git clone git@github.com:your_username/nixopus.git
cd nixopus
```

```bash [HTTPS]
git clone https://github.com/your_username/nixopus.git
cd nixopus
```

:::

::: tip Which Clone Method?
- **SSH**: Requires SSH keys set up with GitHub (faster for frequent pushes)
- **HTTPS**: Works out of the box, may prompt for credentials
:::

### 2. Start Database and SuperTokens

::: warning Port Availability Check
Before starting services, ensure these ports are available on your machine:

| Service | Port | Purpose |
|---------|------|---------|
| PostgreSQL | `5432` | Database |
| SuperTokens | `3567` | Authentication |
| API | `8080` | Backend server |
| Frontend | `3000` | Next.js dev server |

If any port is in use, stop the conflicting service before proceeding.
:::

The docker compose command will automatically start both the PostgreSQL database (`nixopus-db`) and SuperTokens service. The database is a required dependency, so docker compose will start it first and wait for it to be healthy before starting SuperTokens.

::: details Custom Database Credentials
If you need custom database credentials, create a `.env` file in the project root with:

```bash
USERNAME=postgres
PASSWORD=changeme
DB_NAME=postgres
SUPERTOKENS_PORT=3567
```

**Defaults** (used if `.env` is not provided):
- `USERNAME` → `postgres`
- `PASSWORD` → `changeme`
- `DB_NAME` → `postgres`
- `SUPERTOKENS_PORT` → `3567`
:::

Start the services:

```bash
docker compose up supertokens -d
```

Verify containers are running:

```bash
docker ps
```

You should see `nixopus-db` and `supertokens` containers running.

::: tip Container Health
Docker Compose automatically waits for the database to be healthy before starting SuperTokens. This ensures proper initialization order.
:::

### 3. Backend Setup

Navigate to the backend directory:

```bash
cd api
```

Copy the sample environment file:

```bash
cp .env.sample .env
```

::: warning Database Configuration Match
**Critical**: Update the database connection settings in `api/.env` to match your docker compose configuration. The backend connects to the database running in Docker, so these values must match:

- `USERNAME` (must match docker compose)
- `PASSWORD` (must match docker compose)
- `DB_NAME` (must match docker compose)

If they don't match, the backend won't be able to connect to the database.
:::

::: details Missing .env.sample
If `.env.sample` doesn't exist, check the repository structure or create `.env` manually based on the application's configuration requirements.
:::

Download Go dependencies:

```bash
go mod download
```

Install the Air hot reload tool:

```bash
go install github.com/air-verse/air@latest
```

::: tip Air PATH Configuration
Ensure `$GOPATH/bin` or `$HOME/go/bin` is in your `PATH` environment variable so the `air` command is available. Air provides automatic code reloading during development - your Go server will restart automatically when you save changes.
:::

### 4. Frontend Setup

Navigate to the frontend directory:

```bash
cd ../view
```

Copy the sample environment file:

```bash
cp .env.sample .env.local
```

::: details Missing .env.sample
If `.env.sample` doesn't exist, check the repository structure or create `.env.local` manually.
:::

Install Node.js dependencies:

```bash
yarn install
```

::: tip Package Manager
This guide uses `yarn`, but `npm` or `pnpm` will work as well. Use whichever you prefer.
:::

::: info Automatic Git Hooks Setup
When you run `yarn install` (or `npm install`) in the `view` directory, Husky will automatically set up git hooks for commit message validation. This ensures all commits follow the [Conventional Commits](https://www.conventionalcommits.org/) format.

**Commit message format**: `type(scope): description`

**Valid types**: `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `revert`, `style`, `test`

**Examples**:
- ✅ `feat: add user authentication`
- ✅ `fix(api): resolve database connection issue`
- ✅ `docs: update contributing guide`
- ❌ `update code` (missing type)
- ❌ `fix bug` (missing colon)

If your commit message doesn't follow the format, the commit will be rejected with helpful error messages.
:::

## Run

You'll need **two terminal windows** to run both the backend and frontend simultaneously.

### Start Backend

In your **first terminal**, start the backend with Air for hot reloading:

```bash
cd api
air
```

::: info Backend Status
- **URL**: `http://localhost:8080`
- **Hot Reload**: Enabled via Air
- **Auto-restart**: Changes to Go files trigger automatic rebuild and restart
:::

### Start Frontend

In your **second terminal**, start the frontend development server:

```bash
cd view
yarn dev
```

::: info Frontend Status
- **URL**: `http://localhost:3000`
- **Hot Module Replacement**: Enabled
- **Fast Refresh**: Changes appear instantly in the browser
:::

### Verify Everything is Running

Once both servers are running, you should have access to:

| Service | URL | Status Check |
|---------|-----|--------------|
| Frontend | `http://localhost:3000` | Open in browser |
| API | `http://localhost:8080` | Check health endpoint |
| Database | Docker (port 5432) | `docker ps` |
| SuperTokens | Docker (port 3567) | `docker ps` |

::: warning Troubleshooting Connection Issues
If you encounter connection issues, verify:

1. **Docker containers are running**:
   ```bash
   docker ps
   ```
   You should see `nixopus-db` and `supertokens` containers.

2. **Environment variables match**: 
   - Check that `api/.env` database credentials match docker compose configuration

3. **Ports are available**:
   - Ensure ports 3000, 8080, 3567, and 5432 are not blocked or in use

4. **Database is healthy**:
   - Docker Compose waits for health checks, but verify with `docker ps` that containers show as healthy
:::

## Need Help? 🆘

If you run into issues or have questions:

- 💬 [Discord Community](https://discord.gg/skdcq39Wpv) - Get real-time help from the community
- 💡 [GitHub Discussions](https://github.com/raghavyuva/nixopus/discussions) - Ask questions and share ideas

::: tip Contributing
Found a bug or want to improve the docs? Contributions are welcome! Check out the repository for contribution guidelines.
:::

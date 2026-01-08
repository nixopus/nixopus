# Hosting Projects

Deploy applications from GitHub repositories to your VPS with automatic builds, rolling updates, and container management.

## Getting Started

1. Navigate to the **Self Host** section in the dashboard
2. Click **Connect GitHub** to link your GitHub account
3. Select the repository you want to deploy
4. Configure deployment settings (Nixopus auto populates name and description)
5. Click **Deploy**

::: tip
Nixopus automatically fetches available branches from your repository and suggests a default branch.
:::

## Configuration

The deployment wizard guides you through four steps. Here's what each field does:

::: details Step 1: Basic Information
- **Application Name**: Display name for your project
- **Environment**: Choose `Production`, `Staging`, or `Development`
- **Build Pack**: Select `Dockerfile` (default)
- **Port**: The port your application listens on (e.g., `3000`)
:::

::: details Step 2: Repository & Branch
- **Domain**: The domain where your application will be accessible
- **Branch**: Git branch to deploy from (auto detected from your repository)
:::

::: details Step 3: Docker Configuration
- **Build Base Path**: Subdirectory path within your repository where the application code is located. This sets the Docker build context directory. For single-app repositories, use `/` (repository root). For monorepos, specify the subdirectory (e.g., `apps/frontend`). All Docker commands in your Dockerfile will execute relative to this path.

  **Example**: If your repository structure is:
  ```
  my-repo/
    ├── apps/
    │   └── frontend/
    │       ├── src/
    │       └── package.json
    └── README.md
  ```
  Set Build Base Path to `apps/frontend` to use that directory as the build context.

- **Dockerfile Path**: Path to your Dockerfile relative to the build base path. Default is `Dockerfile`. If your Dockerfile is in a subdirectory or has a different name, specify the relative path.

  **Example**: 
  - If Build Base Path is `apps/frontend` and Dockerfile is at `apps/frontend/Dockerfile`, use `Dockerfile`
  - If Build Base Path is `apps/frontend` and Dockerfile is at `apps/frontend/docker/Dockerfile.prod`, use `docker/Dockerfile.prod`
  - If Build Base Path is `/` and Dockerfile is at the root, use `Dockerfile`
:::

::: details Step 4: Variables & Commands
- **Environment Variables**: Runtime variables like `DATABASE_URL=postgres://...`
- **Build Variables**: Build time variables like `NODE_ENV=production`
- **Pre Run Command**: Script that runs before container starts
- **Post Run Command**: Script that runs after container starts
:::

## Monorepo Support

Deploy individual applications from repositories containing multiple projects by setting the **Build Base Path** to your application's directory.

### Understanding Build Base Path and Dockerfile Path

- **Build Base Path**: This is the subdirectory within your repository where your application code lives. Nixopus uses this path as the Docker build context, meaning all Docker commands (like `COPY`, `RUN`, etc.) will execute relative to this directory. For single-app repositories, use `/` (the repository root). For monorepos, specify the subdirectory (e.g., `apps/frontend`).

- **Dockerfile Path**: This is the path to your Dockerfile relative to the build base path. If your Dockerfile is named `Dockerfile` and located in the base path root, use `Dockerfile`. If it's in a subdirectory or has a different name, specify the relative path (e.g., `docker/Dockerfile.prod`).

::: tip How It Works Together
The build process works like this:
1. Nixopus clones your repository
2. Sets the build context to: `repository_root + build_base_path`
3. Looks for the Dockerfile at: `build_context + dockerfile_path`
4. Runs `docker build` with that context and Dockerfile

**Complete Example**:
If your repository structure is:
```
my-monorepo/
  ├── apps/
  │   ├── frontend/
  │   │   ├── docker/
  │   │   │   └── Dockerfile.prod
  │   │   ├── src/
  │   │   └── package.json
  │   └── backend/
  │       ├── Dockerfile
  │       └── src/
  └── README.md
```

For the frontend app:
- **Build Base Path**: `apps/frontend` (this becomes the build context)
- **Dockerfile Path**: `docker/Dockerfile.prod` (relative to the build context)
- **Result**: Docker will build using `apps/frontend` as the context and look for the Dockerfile at `apps/frontend/docker/Dockerfile.prod`

For the backend app:
- **Build Base Path**: `apps/backend`
- **Dockerfile Path**: `Dockerfile` (default, relative to build context)
- **Result**: Docker will build using `apps/backend` as the context and look for the Dockerfile at `apps/backend/Dockerfile`
:::

### Example Structure

```
monorepo/
  ├── apps/
  │   ├── frontend/
  │   │   ├── docker/
  │   │   │   └── Dockerfile.prod
  │   │   └── src/
  │   └── backend/
  │       ├── Dockerfile
  │       └── src/
  └── shared/
      └── libs/
```

::: code-group

```txt [Frontend App]
Build Base Path: apps/frontend
Dockerfile Path: docker/Dockerfile.prod

Build context: monorepo/apps/frontend
Dockerfile location: monorepo/apps/frontend/docker/Dockerfile.prod
```

```txt [Backend App]
Build Base Path: apps/backend
Dockerfile Path: Dockerfile

Build context: monorepo/apps/backend
Dockerfile location: monorepo/apps/backend/Dockerfile
```

:::

## Managing Applications

Click on any deployed application to open its detail view. The application header shows the current status, domain link, environment badge, and any labels you've added.

### Quick Actions

The action buttons in the header let you:

- **Restart** the container without rebuilding (useful for picking up config changes)
- **Redeploy** to rebuild and deploy using the latest code from your branch
- **Redeploy (No Cache)** for a clean build when you need to bypass Docker's layer cache

From the dropdown menu, you can also **Rollback** to a previous deployment or **Delete** the application entirely.

### Organizing with Labels

Add labels to categorize your applications. Click the **Add** button next to the application name, type your label, and press Enter. Labels appear as colored badges and help you identify projects at a glance.

### Monitoring Tab

Track your application's deployment health:

- Total number of deployments
- Success and failure counts
- Current deployment status
- Visual health chart showing trends over time
- Overall success rate percentage

### Logs Tab

View live **Container Logs** to see your application's output in real time. Switch to **Deployment Logs** to review the build process, including clone status, image building, and container startup.

### Configuration Tab

Update your application settings without redeploying:

- Application name and port
- Build base path and Dockerfile path (see Monorepo Support section for examples)
- Environment and build variables
- Pre and post run commands

::: warning
Environment, Branch, Domain, and Build Pack cannot be changed after deployment. To modify these, delete the application and create a new one.
:::

### Deployments Tab

View the complete history of all deployments. Each entry shows the deployment status, timestamp, and commit information. Click on any previous successful deployment to trigger a rollback.

## Automatic Deployments

When you connect a GitHub repository, Nixopus automatically configures webhooks. Every push to your configured branch triggers a new deployment without manual intervention.

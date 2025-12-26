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
- **Base Path**: Root directory for monorepo setups (default: `/`)
- **Dockerfile Path**: Path to Dockerfile relative to base path (default: `/Dockerfile`)
:::

::: details Step 4: Variables & Commands
- **Environment Variables**: Runtime variables like `DATABASE_URL=postgres://...`
- **Build Variables**: Build time variables like `NODE_ENV=production`
- **Pre Run Command**: Script that runs before container starts
- **Post Run Command**: Script that runs after container starts
:::

## Monorepo Support

Deploy individual applications from repositories containing multiple projects by setting the **Base Path** to your application's directory.

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
Base Path: apps/frontend
Dockerfile Path: docker/Dockerfile.prod
```

```txt [Backend App]
Base Path: apps/backend
Dockerfile Path: Dockerfile
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
- Base path and Dockerfile path
- Environment and build variables
- Pre and post run commands

::: warning
Environment, Branch, Domain, and Build Pack cannot be changed after deployment. To modify these, delete the application and create a new one.
:::

### Deployments Tab

View the complete history of all deployments. Each entry shows the deployment status, timestamp, and commit information. Click on any previous successful deployment to trigger a rollback.

## Automatic Deployments

When you connect a GitHub repository, Nixopus automatically configures webhooks. Every push to your configured branch triggers a new deployment without manual intervention.

# Hosting Your Projects
Nixopus makes it easy to host projects on your own VPS. With Nixopus, you can easily manage rolling updates, monitor container statistics, and configure it to run as a CI/CD pipeline.

## Getting Started
To get started, navigate to the self-host section and follow the instructions:

* Connect your GitHub account with Nixopus.
* Choose the project you want to deploy.
* Nixopus will automatically populate the details such as name and description.
* Customize them, if needed.
* Finally, click "Deploy" to start the deployment process.

For more information  on how to deploy projects [visit](#configuration)

## Configuration
| Field | Description | Example |
| --- | --- | --- |
| Port | External port | 3000 |
| Name | Name that describes about your project | My Project |
| Build Pack | Pack to use for building | docker compose / static / dockerfile |
| Environment | Environment type | Dev / Staging / Prod |
| Pre Run Command | Commands to run before starting the container | `npm install` |
| Post Run Command | Commands to run after the container has started | `npm start` |
| Build Variables | Add build variables to your project | `NODE_ENV=production` |
| Environment Variables | Add environment variables to your project | `NODE_ENV=production` |
| Custom Domain | Domain in which your project will be available | `example.com` |
| Base Path | Root directory of your application within the repository (for monorepo setups) | `apps/frontend` |
| Dockerfile Path | Path to Dockerfile relative to the base path | `Dockerfile` or `docker/Dockerfile.prod` |

## Monorepo Support
Nixopus supports deploying applications from monorepo structures. This is particularly useful when you have multiple applications in a single repository.

### Configuration for Monorepos
1. **Base Path**: 
   - Specifies the root directory of your application within the repository
   - Example: If your app is in `apps/frontend`, set base path to `apps/frontend`
   - Default: `/` (root of repository)

2. **Dockerfile Path**:
   - Path to your Dockerfile relative to the base path
   - Example: If Dockerfile is in `apps/frontend/docker/Dockerfile.prod`, set to `docker/Dockerfile.prod`
   - Default: `Dockerfile`

### Example Monorepo Structure
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

### Configuration Examples
1. **Frontend App**:
   - Base Path: `apps/frontend`
   - Dockerfile Path: `docker/Dockerfile.prod`

2. **Backend App**:
   - Base Path: `apps/backend`
   - Dockerfile Path: `Dockerfile` (default)

## Project Management
* Once your project has been deployed, you can manage it from the "Self Host" section.
* You can click on a deployed project to view its details, edit its configuration, redeploy it, or delete it.
* The "Logs" section will show you the container logs and deployment logs.
* The "Deployments" section will show you the information about all the deployments of this project.
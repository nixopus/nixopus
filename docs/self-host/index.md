# Hosting Your Projects

## Introduction
Nixopus makes it easy to host projects on your own VPS. With Nixopus, you can easily manage rolling updates, monitor container statistics, and configure it to run as a CI/CD pipeline.

## Getting Started
To get started, navigate to the self-host section and follow the instructions:

* Connect your GitHub account with Nixopus.
* Choose the project you want to deploy.
* Nixopus will automatically populate the details such as name and description.
* Customize them, if needed.
* Hit "Next".
* Finally, click "Deploy" to start the deployment process.

For more information  on how to deploy projects [visit](#configuration)

## Configuration
| Field | Description | Example |
| --- | --- | --- |
| Ports | External and Internal ports | 5643:5643 |
| Name | Name that describes about your project | My Project |
| Description | Description helps you / your team to find out about the project | A simple todo application written in Golang. |
| Build Pack | Pack to use for building | docker compose / static / dockerfile |
| Environment | Environment type | Dev / Staging / Prod |
| Pre Run Command | Commands to run before starting the container | `npm install` |
| Post Run Command | Commands to run after the container has started | `npm start` |
| Build Variables | Add build variables to your project | `NODE_ENV=production` |
| Environment Variables | Add environment variables to your project | `NODE_ENV=production` |
| Custom Domain | Domain in which your project will be available | `example.com` |

## Project Management
* Once your project has been deployed, you can manage it from the "Self Host" section.
* You can click on a deployed project to view its details, edit its configuration, redeploy it, or delete it.
* The "Logs" section will show you the container logs and deployment logs.
* The "Containers" section will show you the information about all the containers running for this project.
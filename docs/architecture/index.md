# Project Architecture

## Overview

This document provides an overview of the architecture for the project, detailing the various components and their interactions.

```mermaid
flowchart TB
    %% Frontend Layer
    subgraph "Frontend (Next.js)" 
        direction TB
        FE["Next.js Frontend"]:::frontend
        AR["App Router & API Routes"]:::frontend
        CH["Components & Hooks"]:::frontend
        RS["Redux Toolkit Store & Services"]:::frontend
    end

    %% Backend Layer
    subgraph "Backend (Go API Service)"
        direction TB
        BR["Routing & Middleware"]:::backend
        FM["Feature Modules (Auth, Container, Deploy, etc.)"]:::backend
        RT["WebSocket & Realtime Support"]:::backend
    end

    %% Data Store
    DB[(PostgreSQL DB)]:::datastore

    %% Infrastructure
    subgraph "Infrastructure"
        direction TB
        DD["Docker Daemon"]:::infra
        CP["Caddy Reverse Proxy"]:::infra
        DC["Docker Compose Orchestration"]:::infra
        IS["Installer Scripts"]:::infra
        DEV["DevContainer Setup"]:::infra
    end

    %% External Services
    subgraph "External Services"
        direction TB
        GH(("GitHub API")):::external
        NT(("SMTP/Slack/Discord")):::external
    end

    %% Documentation and CI/CD
    DOCS["Documentation Site"]:::external
    CI["GitHub Actions CI/CD"]:::infra

    %% Connections
    Browser["Browser (Client)"]:::frontend
    Browser -->|"HTTPS"| FE
    FE -->|"REST API"| BR
    FE -.->|"WebSocket"| RT
    BR -->|"calls"| FM
    FM -->|"queries"| DB
    RT -->|"LISTEN/NOTIFY"| DB
    BR -->|"Docker SDK"| DD
    BR -->|"Updates Proxy"| CP
    BR -->|"OAuth/Webhooks"| GH
    BR -->|"Notifications"| NT
    IS -->|"runs"| DC
    DC --> DD
    DC --> FE
    DC --> BR
    DC --> DB

    %% Auxiliary
    CP --> FE
    CP --> BR

    %% Documentation & CI/CD placement
    CI --> BR
    DOCS -.-> FE
    DOCS -.-> BR

    %% Click Events
    click FE "https://github.com/raghavyuva/nixopus/tree/master/view/"
    click AR "https://github.com/raghavyuva/nixopus/tree/master/view/app/api/"
    click CH "https://github.com/raghavyuva/nixopus/tree/master/view/components/"
    click CH "https://github.com/raghavyuva/nixopus/tree/master/view/hooks/"
    click RS "https://github.com/raghavyuva/nixopus/tree/master/view/redux/"
    click BR "https://github.com/raghavyuva/nixopus/blob/master/api/internal/routes.go"
    click BR "https://github.com/raghavyuva/nixopus/tree/master/api/internal/middleware/"
    click FM "https://github.com/raghavyuva/nixopus/tree/master/api/internal/features/"
    click RT "https://github.com/raghavyuva/nixopus/tree/master/api/internal/realtime/"
    click DB "https://github.com/raghavyuva/nixopus/tree/master/api/migrations/"
    click DD "https://github.com/raghavyuva/nixopus/blob/master/docker-compose.yml"
    click DD "https://github.com/raghavyuva/nixopus/blob/master/docker-compose-staging.yml"
    click CP "https://github.com/raghavyuva/nixopus/tree/master/helpers/Caddyfile"Add commentMore actions
    click CP "https://github.com/raghavyuva/nixopus/blob/master/helpers/caddy.json"
    click IS "https://github.com/raghavyuva/nixopus/tree/master/installer/"
    click IS "https://github.com/raghavyuva/nixopus/blob/master/scripts/install.sh"
    click DEV "https://github.com/raghavyuva/nixopus/tree/master/.devcontainer/"
    click DOCS "https://github.com/raghavyuva/nixopus/tree/master/docs/"
    click CI "https://github.com/raghavyuva/nixopus/tree/master/.github/workflows/"

    %% Styles
    classDef frontend fill:#D0E8FF,stroke:#333,stroke-width:1px
    classDef backend fill:#D0FFD0,stroke:#333,stroke-width:1px
    classDef datastore fill:#FFE0A0,stroke:#333,stroke-width:1px
    classDef external fill:#E0E0E0,stroke:#333,stroke-width:1px
    classDef infra fill:#F0F0F0,stroke:#333,stroke-width:1px
```

## API Layer

- **Language**: Go
- **Location**: `api` directory
- **Description**: The API layer is built using Go, providing backend services. It includes a `Dockerfile` for containerization and uses Go modules for dependency management.

## Frontend Layer

- **Framework**: Likely Next.js
- **Location**: `view` directory
- **Description**: The frontend is built using a JavaScript framework, with configuration files indicating the use of Next.js. It includes a `Dockerfile` for containerization and a `package.json` for managing dependencies.

## Assets

- **Location**: `assets` directory
- **Description**: Contains image files used for branding and UI elements.

## Installation

- **Language**: Python
- **Location**: `installer` directory
- **Description**: Python scripts are used for installation and setup, managing the installation process and dependencies.

## Deployment

- **Tools**: Docker, Docker Compose
- **Description**: The project uses Docker for containerization and Docker Compose for orchestrating multi-container applications. Configuration files are provided for both development and staging environments.

## Web Server

- **Tool**: Caddy
- **Location**: `helpers` directory
- **Description**: Caddy is used as a web server, with configuration files provided for setup.

## Documentation

- **Location**: `docs` directory
- **Description**: Contains various documentation files and directories, including this architecture overview.

_**Nixopus**_ is an end-to-end platform with a UI built on React.js. The backend is written in Go, which is responsible for managing core functions, including authentication, deployments, and real-time updates via WebSockets.

Data Storage and Management is supported via PostgreSQL database, which the backend accesses for queries and real-time notifications. The entire system operates within Docker containers, coordinated using Docker Compose, with Caddy serving as a reverse proxy to route requests securely.

Nixopus integrates with external services, currently GitHub for login and webhooks, and uses email, Slack, and Discord to send notifications.

Installer scripts are purely used for self-hosting and production deployment purposes only, and development container setups are available to help developers quickly start working on the project. Automated CI/CD pipelines handle testing and deployment, while built-in documentation supports easy maintenance.

This architecture setup allows for a modular and scalable application, leveraging containerization for easy deployment and management.

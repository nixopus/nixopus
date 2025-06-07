# Architecture

## Overview

This document provides an overview of the architecture for the project, detailing the various components and their interactions.

```mermaid
%%{init: {
  'themeVariables': {
    'fontSize': '16px',
    'primaryColor': '#87ceeb',
    'edgeLabelBackground':'#0067ff'
  },
  'flowchart': {
    'zoom': true
  }
}}%%

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
    click RS "https://github.com/raghavyuva/nixopus/tree/master/view/redux/"
    click BR "https://github.com/raghavyuva/nixopus/tree/master/api/internal/middleware/"
    click FM "https://github.com/raghavyuva/nixopus/tree/master/api/internal/features/"
    click RT "https://github.com/raghavyuva/nixopus/tree/master/api/internal/realtime/"
    click DB "https://github.com/raghavyuva/nixopus/tree/master/api/migrations/"
    click DD "https://github.com/raghavyuva/nixopus/blob/master/docker-compose.yml"
    click CP "https://github.com/raghavyuva/nixopus/tree/master/helpers/Caddyfile"
    click IS "https://github.com/raghavyuva/nixopus/tree/master/installer/"
    click DEV "https://github.com/raghavyuva/nixopus/tree/master/.devcontainer/"
    click DOCS "https://github.com/raghavyuva/nixopus/tree/master/docs/"
    click CI "https://github.com/raghavyuva/nixopus/tree/master/.github/workflows/"

    %% Styles (higher-contrast palettes)
    classDef frontend fill:#4682B4,stroke:#333,stroke-width:1px,color:#fff;
    classDef backend fill:#32CD32,stroke:#333,stroke-width:1px,color:#fff;
    classDef datastore fill:#FFD700,stroke:#333,stroke-width:1px,color:#000;
    classDef external fill:#b000000,stroke:#333,stroke-width:1px,color:#fff;
    classDef infra fill:#CD5C5C,stroke:#333,stroke-width:1px,color:#fff;

```

## API Layer

- **Language**: Go
- **Location**: [api](https://github.com/raghavyuva/nixopus/tree/master/api) directory
- **Description**: The API layer is built using Go, providing backend services. It includes a [Dockerfile](https://github.com/raghavyuva/nixopus/blob/master/api/Dockerfile) for containerization and uses Go modules for dependency management.

## Frontend Layer

- **Framework**: Next.js
- **Location**: [view](https://github.com/raghavyuva/nixopus/tree/master/view) directory
- **Description**: The frontend is built using a JavaScript framework, with configuration files indicating the use of Next.js. It includes a [Dockerfile](https://github.com/raghavyuva/nixopus/blob/master/view/Dockerfile) for containerization and a [package.json](https://github.com/raghavyuva/nixopus/blob/master/view/package.json) for managing dependencies.

## Assets

- **Location**: [assets](https://github.com/raghavyuva/nixopus/tree/master/assets) directory
- **Description**: Contains image files used for branding and UI elements.

## Installation

- **Language**: Python
- **Location**: [installer](https://github.com/raghavyuva/nixopus/tree/master/installer) directory
- **Description**: Python scripts are used for installation and setup, managing the installation process and dependencies.

## Deployment

- **Tools**: `Docker`, `Docker Compose`
- **Description**: The project uses `Docker` for containerization and `Docker Compose` for orchestrating multi-container applications. Configuration files are provided for both development and staging environments.

## Web Server

- **Tool**: Caddy
- **Location**: [helpers](https://github.com/raghavyuva/nixopus/tree/master/helpers) directory
- **Description**: Caddy is used as a web server, with configuration files provided for setup.

## Documentation

- **Location**: [docs](https://github.com/raghavyuva/nixopus/tree/master/docs) directory
- **Description**: Contains various documentation files and directories, including this architecture overview.

## TLDR

> Note: The architecture overview provided is a high-level representation of the project's structure and components.

- UI: React.js (Next.js)
- Backend: Go (authentication, deployments, real-time via WebSockets)
- Data Storage: PostgreSQL (uses LISTEN/NOTIFY for realtime)
- Containerization: Docker Compose (each component in its own container)
- Reverse Proxy: Caddy (routes HTTPS, manages TLS)
- Integrations:
  - GitHub (OAuth & webhooks)
  - Notifications via Email, Slack, Discord
  - Setup & Dev: .devcontainer/ for local development
  - Production: Python installer scripts for production
- TestingCI/CD: GitHub Actions (automated testing and deployments)
- Documentation: Built-in docs under docs/ using vitepress

This architecture setup allows for a modular and scalable application, leveraging containerization for easy deployment and management.

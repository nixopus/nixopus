---
name: shell-deploy
description: Deploy shell script applications — interpreter detection, setup scripts, and Dockerfile patterns. Use when deploying a shell script project, or when start.sh is detected.
metadata:
  version: "1.0"
---

# Shell Script Deployment

## Detection

Project is a shell script application if `start.sh` exists in the root directory.

## Shell Interpreter

Detected from the shebang line of the entry script:

| Shell | Shebang | Notes |
|---|---|---|
| bash | `#!/bin/bash` | Available in base image |
| sh | `#!/bin/sh` | Available in base image |
| dash | `#!/bin/dash` | Uses sh in base image |
| zsh | `#!/bin/zsh` | Must be installed in Dockerfile |

No shebang defaults to `sh`.

## Setup Scripts

If a `setup.sh` file exists alongside `start.sh`, it runs during the Docker build to prepare the environment (install tools, generate config, etc.). Files created by the setup script are available at runtime.

## Dockerfile Patterns

### Basic

```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY start.sh .
RUN chmod +x start.sh
EXPOSE 8080
CMD ["./start.sh"]
```

### With setup script

```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . .
RUN chmod +x start.sh setup.sh
RUN ./setup.sh
EXPOSE 8080
CMD ["./start.sh"]
```

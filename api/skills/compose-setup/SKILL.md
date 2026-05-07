---
name: compose-setup
description: Generate docker-compose.yml for multi-service setups including databases, caches, and service dependencies. Use when the app needs a database, cache, message broker, or has multiple independently deployable services.
metadata:
  version: "1.0"
---

# Docker Compose Setup

Generate `docker-compose.yml` when the application needs multiple services (app + database, app + cache, monorepo services).

## When to use docker-compose

- App requires a database (PostgreSQL, MySQL, MongoDB, Redis)
- App has multiple independently deployable services
- App needs a message broker (RabbitMQ, Kafka)
- Monorepo with multiple `Dockerfile` entries

Single-service apps with no external dependencies should use a standalone Dockerfile.

## Detection from source code

Look for these signals to determine required services:

| Pattern in code or config | Service needed |
|--------------------------|----------------|
| `DATABASE_URL=postgres://` or `pg` dependency | PostgreSQL |
| `DATABASE_URL=mysql://` or `mysql2` dependency | MySQL |
| `MONGODB_URI` or `mongoose` dependency | MongoDB |
| `REDIS_URL` or `ioredis`/`redis` dependency | Redis |
| `RABBITMQ_URL` or `amqplib` dependency | RabbitMQ |
| `KAFKA_BROKERS` | Kafka |
| `.env.example` listing service URLs | Parse each URL for service type |

## Compose template structure

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "${PORT:-3000}:3000"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/app
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: app
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
```

## Common service blocks

### PostgreSQL

```yaml
db:
  image: postgres:16-alpine
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: app
  volumes:
    - pgdata:/var/lib/postgresql/data
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U postgres"]
    interval: 5s
    timeout: 3s
    retries: 5
```

### MySQL

```yaml
db:
  image: mysql:8.0
  environment:
    MYSQL_ROOT_PASSWORD: root
    MYSQL_DATABASE: app
    MYSQL_USER: app
    MYSQL_PASSWORD: app
  volumes:
    - mysqldata:/var/lib/mysql
  healthcheck:
    test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
    interval: 5s
    timeout: 3s
    retries: 5
```

### MongoDB

```yaml
mongo:
  image: mongo:7
  environment:
    MONGO_INITDB_ROOT_USERNAME: root
    MONGO_INITDB_ROOT_PASSWORD: root
  volumes:
    - mongodata:/data/db
  healthcheck:
    test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
    interval: 5s
    timeout: 3s
    retries: 5
```

### Redis

```yaml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes
  volumes:
    - redisdata:/data
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s
    timeout: 3s
    retries: 5
```

## Monorepo services

For monorepos with multiple services under `apps/` or `services/`:

```yaml
services:
  api:
    build:
      context: .
      dockerfile: apps/api/Dockerfile
    ports:
      - "3001:3001"
    depends_on:
      db:
        condition: service_healthy

  web:
    build:
      context: .
      dockerfile: apps/web/Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      - api
```

Set build context to the repo root so shared packages are accessible. Each Dockerfile handles its own `COPY` paths.

## Rules

- Always use named volumes for persistent data
- Always add healthchecks for databases and caches
- Use `depends_on` with `condition: service_healthy` so the app waits for healthy dependencies
- Use `restart: unless-stopped` for all services
- Reference env vars from `.env.example` when available
- Default to Alpine/slim images for smaller footprint
- Pin image versions (e.g. `postgres:16-alpine`, not `postgres:latest`)

## Related Skills

- **`env-detection`** — Detect required service connection URLs (DATABASE_URL, REDIS_URL, etc.) to determine which services to include
- **`dockerfile-generation`** — Generate the Dockerfile for each service in the compose file

---
name: database-migration
description: Run database migrations safely during deployment — framework-specific commands, pre-deploy vs post-deploy timing, health gates, and rollback strategies. Use when the app has a database migration system and needs migrations run during deployment.
metadata:
  version: "1.0"
---

# Database Migration

## Detection

Check for migration tooling in the project:

| Signal | Migration tool | Ecosystem |
|---|---|---|
| `prisma/schema.prisma` or `@prisma/client` in deps | Prisma | Node.js |
| `typeorm` in deps + `ormconfig` or `data-source.ts` | TypeORM | Node.js |
| `knex` in deps + `knexfile` | Knex | Node.js |
| `drizzle-orm` in deps + `drizzle.config.ts` | Drizzle | Node.js |
| `sequelize` in deps + `config/config.json` | Sequelize | Node.js |
| `manage.py` + Django in deps | Django | Python |
| `alembic/` directory or `alembic` in deps | Alembic | Python |
| `flask-migrate` in deps | Flask-Migrate | Python |
| `goose` or `migrate` in go.mod | Goose / golang-migrate | Go |
| `ActiveRecord` + `db/migrate/` | Rails Migrations | Ruby |
| `ecto` in mix.exs | Ecto | Elixir |
| `flyway` or `liquibase` in pom.xml / build.gradle | Flyway / Liquibase | Java |
| `Entity Framework` in .csproj | EF Core | .NET |

## Migration Commands

| Tool | Migrate command | Status/check command |
|---|---|---|
| Prisma | `npx prisma migrate deploy` | `npx prisma migrate status` |
| TypeORM | `npx typeorm migration:run` | `npx typeorm migration:show` |
| Knex | `npx knex migrate:latest` | `npx knex migrate:status` |
| Drizzle | `npx drizzle-kit migrate` | `npx drizzle-kit check` |
| Sequelize | `npx sequelize-cli db:migrate` | `npx sequelize-cli db:migrate:status` |
| Django | `python manage.py migrate` | `python manage.py showmigrations` |
| Alembic | `alembic upgrade head` | `alembic current` |
| Flask-Migrate | `flask db upgrade` | `flask db current` |
| Goose | `goose up` | `goose status` |
| golang-migrate | `migrate -path ./migrations -database $DATABASE_URL up` | `migrate ... version` |
| Rails | `bundle exec rake db:migrate` | `bundle exec rake db:migrate:status` |
| Ecto | `mix ecto.migrate` | `mix ecto.migrations` |
| Flyway | `flyway migrate` | `flyway info` |
| Liquibase | `liquibase update` | `liquibase status` |
| EF Core | `dotnet ef database update` | `dotnet ef migrations list` |

## When to Run Migrations

### Pre-deploy (before new code runs)

Use when: new code REQUIRES the schema change to function.

- Run migration as a separate step before deploying the new container
- If migration fails, abort deployment — don't start the new container
- Compose: use a `migrate` service with `depends_on` before the app service

### Post-deploy (as part of container startup)

Use when: migration is additive (new columns/tables) and old code wouldn't break.

- Include migration command in Dockerfile CMD or entrypoint script
- Risk: if migration fails, the container may crash-loop
- Advantage: simpler deployment pipeline

### Recommended patterns by framework

| Framework | Pattern | Implementation |
|---|---|---|
| Prisma | Entrypoint script | `npx prisma migrate deploy && node dist/index.js` |
| Django | Entrypoint script | `python manage.py migrate && gunicorn ...` |
| Rails | Entrypoint script | `bundle exec rake db:migrate && bundle exec puma ...` |
| Alembic | Pre-deploy step | Run `alembic upgrade head` before deploying |
| Ecto | Release command | `mix ecto.migrate` as release pre-start hook |
| EF Core | Pre-deploy step | `dotnet ef database update` before deploying |

## Compose Migration Service

For compose deployments, add a migration service that runs before the app:

```yaml
services:
  migrate:
    build: .
    command: npx prisma migrate deploy
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/app
    depends_on:
      db:
        condition: service_healthy

  app:
    build: .
    depends_on:
      migrate:
        condition: service_completed_successfully
      db:
        condition: service_healthy
```

## Entrypoint Script Pattern

When migrations run at container startup:

```bash
#!/bin/sh
set -e

echo "Running migrations..."
npx prisma migrate deploy

echo "Starting application..."
exec node dist/index.js
```

Key: use `exec` for the final command so the app process becomes PID 1 and receives signals correctly.

## Safe Migration Practices

- **Always use `migrate deploy` / `migrate:latest`** (not `push` or `sync`) — deploy applies migration files in order; push/sync can be destructive
- **Never run migrations interactively** — all migration commands must work non-interactively in Docker
- **DATABASE_URL must be set** — migrations need the production database connection, not a build-time placeholder
- **Additive-first**: add new columns as nullable or with defaults before deploying code that requires them
- **Separate schema changes from data changes** — schema migrations in deploy pipeline, data backfills as separate tasks
- **Test migrations against a copy** before running on production when possible

## Gotchas

- Prisma `migrate deploy` vs `db push`: `deploy` applies migration files; `push` syncs schema directly (destructive, dev-only)
- Django `migrate` with `--run-syncdb` can create tables without migration files — avoid in production
- TypeORM `synchronize: true` in production drops and recreates tables — ensure it's disabled
- Alembic `autogenerate` may miss some changes (custom types, triggers) — always review generated migrations
- Rails `db:schema:load` vs `db:migrate`: `schema:load` replaces all migrations with a single schema load — only use for new databases
- EF Core `Update-Database` in Package Manager Console is interactive — use `dotnet ef database update` for Docker

## Related Skills

- **`pre-deploy-checklist`** — Detects migration tools and checks if migration command is in the deploy flow
- **`rollback-strategy`** — Guidance on rolling back when migrations make rollback risky
- **`compose-setup`** — Migration service pattern for compose deployments

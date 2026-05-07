---
name: php-deploy
description: Build and deploy PHP applications — Composer, Laravel, Symfony, FrankenPHP, PHP-FPM, and Dockerfile patterns. Use when deploying a PHP project, or when composer.json or index.php is detected.
metadata:
  version: "1.0"
---

# PHP Deployment

## Detection

Project is PHP if any of these exist:

- `composer.json` at the build context root
- `index.php` at the root

## Versions

PHP 8.2+ only.

1. `composer.json` → `require.php` (e.g. `">=8.2"`)
2. `.php-version` or `mise.toml` / `.tool-versions`
3. Defaults to **8.4**

## Package Manager

Composer. Detect `composer.json` and `composer.lock`.

| Has lock file | Command |
|---|---|
| yes | `composer install --no-dev --optimize-autoloader` |
| no | `composer install --no-dev --optimize-autoloader` |

## Document Root

Laravel and Symfony use `public/`. Generic PHP uses the project root.

## PHP Extensions

Detected from `composer.json` → `require` section with `ext-*` entries.

## Framework Detection

| Signal | Framework |
|---|---|
| `artisan` | Laravel |
| `symfony.lock` or `config/` structure | Symfony |
| `wp-config.php` | WordPress |
| `index.php` + routing | Generic PHP |

### Laravel-Specific

When `artisan` is present:

- Document root: `public/`
- Migrations run by default during startup. Skip by setting `SKIP_MIGRATIONS=true`.
- Storage symlinks, optimization
- Build-time artisan caches: config, routes, views, events
- Storage directory must be writable

## Startup Process

1. Run migrations (Laravel; skip if `SKIP_MIGRATIONS=true`)
2. Create storage symlinks (Laravel: `php artisan storage:link`)
3. Optimize (Laravel: config/route/view cache)
4. Start PHP server (FrankenPHP, PHP-FPM, or `php -S`)

Override by providing custom `start-container.sh` in project root.

## Node.js Integration

If `package.json` detected alongside PHP (e.g. Laravel + Vite):

- Install Node.js
- Run `npm install` or equivalent
- Execute build scripts from `package.json`
- Prune dev dependencies in final image

## Install Stage Optimization

Copy `composer.json` + `composer.lock` first for layer caching.

## Caching

```dockerfile
RUN --mount=type=cache,target=/root/.composer/cache \
    composer install --no-dev --optimize-autoloader
```

## Base Images

| Use case | Image |
|---|---|
| FrankenPHP | `dunglas/frankenphp:latest` or `dunglas/frankenphp:php8.4` |
| PHP-FPM | `php:8.4-fpm-alpine` |
| CLI / built-in server | `php:8.4-cli-alpine` |

## Dockerfile Patterns

### Laravel with FrankenPHP

```dockerfile
FROM composer:2 AS deps
WORKDIR /app
COPY composer.json composer.lock ./
RUN composer install --no-dev --optimize-autoloader

FROM php:8.4-cli-alpine AS build
WORKDIR /app
COPY --from=deps /app/vendor ./vendor
COPY . .
RUN php artisan config:cache && php artisan route:cache && php artisan view:cache

FROM dunglas/frankenphp:php8.4
WORKDIR /app
COPY --from=build /app .
EXPOSE 8080
CMD ["frankenphp", "run", "--config", "/app/Caddyfile"]
```

### Generic PHP + Composer

```dockerfile
FROM composer:2 AS deps
WORKDIR /app
COPY composer.json composer.lock ./
RUN composer install --no-dev --optimize-autoloader

FROM php:8.4-fpm-alpine
WORKDIR /app
COPY --from=deps /app/vendor ./vendor
COPY . .
EXPOSE 9000
CMD ["php-fpm"]
```

## Gotchas

- Laravel `storage/` and `bootstrap/cache/` must be writable — `chown` these directories in the Dockerfile for non-root users
- FrankenPHP defaults to port 8080, not 80 — ensure EXPOSE and proxy config match
- Composer `post-autoload-dump` scripts may fail during Docker build if the app isn't fully configured — use `--no-scripts` during install, then run `composer dump-autoload --optimize` after copying source
- PHP OPcache should be enabled in production (`opcache.enable=1`, `opcache.validate_timestamps=0`) for significant performance gains
- Extensions listed in `require` as `ext-*` must be installed in the Docker image — they are not bundled with base PHP images

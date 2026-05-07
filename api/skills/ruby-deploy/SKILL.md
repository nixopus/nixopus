---
name: ruby-deploy
description: Build and deploy Ruby applications — Bundler, Rails, Rack, asset pipeline, bootsnap, and Dockerfile patterns. Use when deploying a Ruby or Rails project, or when a Gemfile is detected.
metadata:
  version: "1.0"
---

# Ruby Deployment

## Detection

Project is Ruby if `Gemfile` exists at the root.

## Versions

Ruby version priority:

1. `.ruby-version` file
2. `Gemfile` → `ruby` directive
3. `mise.toml` or `.tool-versions`
4. Defaults to **3.4**

## Package Manager

Bundler. Detect `Gemfile` and `Gemfile.lock`.

### Bundler Version

Read from `BUNDLED WITH` in `Gemfile.lock`. Install: `gem install bundler -v <version>`.

### Install Commands

| Context | Command |
|---|---|
| Production | `bundle config set --local without 'development test'` then `bundle install` |
| With lockfile | `bundle install --frozen` |
| Without lockfile | `bundle install` |

## Runtime Variables

```
BUNDLE_GEMFILE=/app/Gemfile
GEM_PATH=/usr/local/bundle
GEM_HOME=/usr/local/bundle
MALLOC_ARENA_MAX=2
```

## Build & Start

### Start Command Resolution

1. Framework-specific (see below)
2. `config/environment.rb` → load and run
3. `config.ru` → Rack: `bundle exec rackup config.ru`
4. `Rakefile` → `bundle exec rake <task>`

## Framework Detection

| Signal | Framework |
|---|---|
| `config/application.rb` | Rails |
| `config.ru` | Rack |
| `Rakefile` | Rake-based app |

### Rails-Specific

When `config/application.rb` is present:

- Asset precompile: `bundle exec rake assets:precompile` (Sprockets or Propshaft)
- API-only Rails (no asset pipeline gems) skip asset compilation
- Bootsnap: `bundle exec bootsnap precompile --gemfile` during deps; `bundle exec bootsnap precompile app/ lib/` during build

## System Dependencies

| Gem / use | System dependency |
|---|---|
| `pg` | `libpq-dev` |
| `mysql2` | `default-libmysqlclient-dev` |
| `rmagick` | `libmagickwand-dev` |
| `image_processing` (vips) | `libvips-dev` |
| `charlock_holmes` | `libicu-dev`, `libxml2-dev`, `libxslt-dev` |

## Asset Pipeline

| Gem | Build command |
|---|---|
| Sprockets | `bundle exec rake assets:precompile` |
| Propshaft | `bundle exec rake assets:precompile` |

Rails API-only (no `sprockets-rails`, `propshaft`) skip asset compilation.

## Bootsnap

When `bootsnap` is in Gemfile:

- During deps: `bundle exec bootsnap precompile --gemfile`
- During build (Rails): `bundle exec bootsnap precompile app/ lib/`

## Node.js Integration

If `package.json` detected, or `execjs` gem is used:

- Install Node.js
- Run `npm install` (or equivalent)
- Execute build scripts
- Prune dev dependencies

Common for Rails with Webpacker, Vite, or other frontend tooling.

## Performance Optimizations

| Optimization | When |
|---|---|
| jemalloc | Install `libjemalloc2` for improved memory allocation |
| YJIT | Ruby 3.2+: may need `rustc` and `cargo` for compilation |

## Install Stage Optimization

Copy `Gemfile` + `Gemfile.lock` first for layer caching.

## Caching

```dockerfile
RUN --mount=type=cache,target=/usr/local/bundle \
    bundle config set --local path /usr/local/bundle \
    && bundle install --frozen
```

## Base Images

| Stage | Image |
|---|---|
| Build | `ruby:3.4-bookworm` or `ruby:3.4-alpine` |
| Runtime | `ruby:3.4-slim` or `ruby:3.4-alpine` |

## Dockerfile Patterns

### Rails

```dockerfile
FROM ruby:3.4-bookworm AS base

FROM base AS deps
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle config set --local path /usr/local/bundle \
    && bundle install --frozen

FROM base AS build
WORKDIR /app
COPY --from=deps /usr/local/bundle /usr/local/bundle
COPY . .
ENV RAILS_ENV=production
RUN bundle exec bootsnap precompile app/ lib/ \
    && bundle exec rake assets:precompile

FROM ruby:3.4-slim
WORKDIR /app
ENV RAILS_ENV=production
ENV BUNDLE_GEMFILE=/app/Gemfile
ENV GEM_HOME=/usr/local/bundle
ENV GEM_PATH=/usr/local/bundle
COPY --from=deps /usr/local/bundle /usr/local/bundle
COPY --from=build /app/public/assets /app/public/assets
COPY . .
EXPOSE 3000
CMD ["bundle", "exec", "puma", "-C", "config/puma.rb"]
```

### Rack (config.ru)

```dockerfile
FROM ruby:3.4-slim
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle config set --local path /usr/local/bundle \
    && bundle install --frozen
COPY . .
EXPOSE 9292
CMD ["bundle", "exec", "rackup", "config.ru", "-o", "0.0.0.0", "-p", "9292"]
```

## Gotchas

- Bundler version mismatch is a top build failure — always install the exact version from `BUNDLED WITH` in `Gemfile.lock`
- Rails `SECRET_KEY_BASE` must be set even for asset precompilation — use a dummy value at build time: `SECRET_KEY_BASE=dummy bundle exec rake assets:precompile`
- `bundle exec` is required in Docker CMD — running `puma` directly may use a system gem instead of the bundled one
- jemalloc can reduce Rails memory usage 30-40% — `LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libjemalloc.so.2` or install `libjemalloc2`
- Bootsnap precompilation caches must happen AFTER gems are installed — running `bootsnap precompile --gemfile` before `bundle install` will fail silently

---
name: elixir-deploy
description: Build and deploy Elixir and Phoenix applications — version detection, mix releases, and Dockerfile patterns. Use when deploying an Elixir or Phoenix project, or when mix.exs is detected.
metadata:
  version: "1.0"
---

# Elixir Deployment

## Detection

Project is Elixir if `mix.exs` exists in the root.

## Versions

### Elixir

1. `.elixir-version` file
2. `mix.exs` → detected from project
3. `mise.toml` or `.tool-versions`
4. Defaults to **1.18**

### Erlang (OTP)

1. `.erlang-version` file
2. Resolved automatically from Elixir version
3. `mise.toml` or `.tool-versions`
4. Defaults to **27.3**

## Build

### Build Process

1. Install Elixir and Erlang
2. Get and compile deps: `mix deps.get --only prod` and `mix deps.compile`
3. If defined: `mix assets.setup`
4. If defined: `mix assets.deploy` and `mix ecto.deploy`
5. Compile and release: `mix compile` and `mix release`

### Start Command

```
/app/_build/prod/rel/<app_name>/bin/<app_name> start
```

## Framework Detection

| Signal | Framework |
|---|---|
| `mix.exs` + `phoenix` dep | Phoenix |
| `config/config.exs` | Standard Elixir app |
| `config/runtime.exs` | Runtime config (prod) |

### Phoenix-Specific

- Asset pipeline: Node.js or esbuild; run `mix assets.setup` and `mix assets.deploy`
- Ecto: `mix ecto.setup` (dev) / `mix ecto.deploy` (prod)
- Release: `mix release` produces standalone tarball/binary

## Install Stage Optimization

Copy in order:
- `mix.exs`, `mix.lock`
- `config/` (full config directory)
- `lib/`, `test/`, `assets/` (Phoenix)

## Base Images

| Stage | Image |
|---|---|
| Build | `hexpm/elixir:1.18-erlang-27.3` or `elixir:1.18` |
| Runtime | `debian:bookworm-slim` + extracted release, or `elixir:1.18-slim` |

## Dockerfile Patterns

### Phoenix Release

```dockerfile
FROM hexpm/elixir:1.18-erlang-27.3 AS build
WORKDIR /app
RUN mix local.hex --force && mix local.rebar --force
COPY mix.exs mix.lock ./
RUN mix deps.get --only prod && mix deps.compile
COPY config config
COPY lib lib
COPY priv priv
COPY assets assets
RUN mix assets.deploy
RUN mix compile
RUN mix release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libssl3 ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/_build/prod/rel/my_app ./
EXPOSE 4000
CMD ["./bin/my_app", "start"]
```

### Plain Elixir (no Phoenix)

```dockerfile
FROM hexpm/elixir:1.18-erlang-27.3 AS build
WORKDIR /app
COPY mix.exs mix.lock ./
RUN mix deps.get --only prod && mix deps.compile
COPY config config
COPY lib lib
RUN mix compile && mix release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libssl3 ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/_build/prod/rel/my_app ./
CMD ["./bin/my_app", "start"]
```

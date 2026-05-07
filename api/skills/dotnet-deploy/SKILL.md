---
name: dotnet-deploy
description: Build and deploy .NET applications — ASP.NET Core, version detection, self-contained builds, and Dockerfile patterns. Use when deploying a .NET or C# project, or when a .csproj file is detected.
metadata:
  version: "1.0"
---

# .NET Deployment

## Detection

Project is .NET if a `*.csproj` file exists in the root.

## Versions

1. `*.csproj` → `TargetFramework` (e.g. `net8.0`)
2. `global.json` → `sdk.version`
3. Defaults to **8.0**

## Build

### Build Process

1. Restore: `dotnet restore`
2. Publish: `dotnet publish --no-restore -c Release -o out`
3. Output: `./out/<project_name>.dll` or `./out/<project_name>.exe`

### Start Command

- `./out/<project_name>` or `dotnet ./out/<project_name>.dll`
- Project name from `*.csproj` filename (e.g. `MyApp.csproj` → `MyApp`)

## Port Binding

```
ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000}
```

Ensures the app listens on all interfaces.

## Runtime Packages

| Package | Purpose |
|---|---|
| `libicu-dev` | Internationalization support |

## Framework Detection

| Signal | Framework |
|---|---|
| `Microsoft.AspNetCore.App` | ASP.NET Core |
| `Microsoft.NET.Sdk.Web` | Web SDK |
| `Microsoft.NET.Sdk.Worker` | Worker service |
| `Microsoft.NET.Sdk` | Console / class library |

## Install Stage Optimization

Copy project files first for layer caching:
- `*.csproj`, `*.sln`
- `global.json`, `nuget.config`, `Directory.Build.props`, `Directory.Packages.props`
- Then `src/`

## Caching

```dockerfile
RUN --mount=type=cache,target=/root/.nuget/packages \
    dotnet restore
```

## Base Images

| Stage | Image |
|---|---|
| Build | `mcr.microsoft.com/dotnet/sdk:8.0` or `mcr.microsoft.com/dotnet/sdk:8.0-alpine` |
| Runtime | `mcr.microsoft.com/dotnet/aspnet:8.0` or `mcr.microsoft.com/dotnet/runtime:8.0` |

Use ASP.NET runtime for web apps; `runtime` for console/worker.

## Dockerfile Patterns

### ASP.NET Core

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app
COPY *.csproj ./
RUN dotnet restore
COPY . .
RUN dotnet publish --no-restore -c Release -o out

FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=build /app/out .
ENV ASPNETCORE_URLS=http://0.0.0.0:3000
EXPOSE 3000
ENTRYPOINT ["dotnet", "MyApp.dll"]
```

### Multi-project (sln)

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app
COPY *.sln ./
COPY src/ ./src/
RUN dotnet restore
RUN dotnet publish src/MyApp/MyApp.csproj --no-restore -c Release -o out

FROM mcr.microsoft.com/dotnet/aspnet:8.0
WORKDIR /app
COPY --from=build /app/out .
ENV ASPNETCORE_URLS=http://0.0.0.0:3000
EXPOSE 3000
ENTRYPOINT ["dotnet", "MyApp.dll"]
```

### Self-contained (single binary)

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app
COPY *.csproj ./
RUN dotnet restore
COPY . .
RUN dotnet publish --no-restore -c Release -o out -r linux-x64 --self-contained true -p:PublishSingleFile=true

FROM mcr.microsoft.com/dotnet/runtime-deps:8.0
WORKDIR /app
COPY --from=build /app/out .
ENV ASPNETCORE_URLS=http://0.0.0.0:3000
EXPOSE 3000
ENTRYPOINT ["./MyApp"]
```

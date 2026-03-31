# Acknowledgments

Nixopus is built on the shoulders of outstanding open-source projects, tools, and services. This file recognizes every project, library, and resource that helped make Nixopus possible.

---

## Core Technologies

| Project | Description | License |
|---------|-------------|---------|
| [Go](https://go.dev) | Backend language | BSD-3-Clause |
| [Next.js](https://nextjs.org) | React framework powering the dashboard | MIT |
| [React](https://react.dev) | UI library | MIT |
| [TypeScript](https://www.typescriptlang.org) | Type-safe JavaScript | Apache-2.0 |
| [Docker](https://www.docker.com) | Containerization engine | Apache-2.0 |
| [PostgreSQL](https://www.postgresql.org) | Primary database | PostgreSQL License |
| [Redis](https://redis.io) | In-memory data store | RSALv2 / SSPLv1 |

---

## Infrastructure & Services

| Project | Role in Nixopus | License |
|---------|-----------------|---------|
| [Caddy](https://caddyserver.com) | Reverse proxy and automatic TLS | Apache-2.0 |
| [Let's Encrypt](https://letsencrypt.org) | Free, automated SSL certificates | — |
| [Docker Compose](https://docs.docker.com/compose/) | Multi-container orchestration | Apache-2.0 |

---

## Go Libraries

| Library | Purpose |
|---------|---------|
| [go-fuego/fuego](https://github.com/go-fuego/fuego) | HTTP framework |
| [jackc/pgx](https://github.com/jackc/pgx) | PostgreSQL driver |
| [uptrace/bun](https://github.com/uptrace/bun) | SQL-first ORM |
| [docker/docker](https://github.com/moby/moby) | Docker Engine client |
| [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket connections |
| [caddyserver/caddy](https://github.com/caddyserver/caddy) | Caddy library integration |
| [raghavyuva/caddygo](https://github.com/raghavyuva/caddygo) | Caddy Go bindings |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | JWT authentication |
| [lestrrat-go/jwx](https://github.com/lestrrat-go/jwx) | JWK / JWS / JWE utilities |
| [go-playground/validator](https://github.com/go-playground/validator) | Struct validation |
| [go-redis/redis](https://github.com/go-redis/redis) | Redis client |
| [sirupsen/logrus](https://github.com/sirupsen/logrus) | Structured logging |
| [spf13/viper](https://github.com/spf13/viper) | Configuration management |
| [robfig/cron](https://github.com/robfig/cron) | Cron scheduler |
| [joho/godotenv](https://github.com/joho/godotenv) | `.env` file loading |
| [google/uuid](https://github.com/google/uuid) | UUID generation |
| [pkg/sftp](https://github.com/pkg/sftp) | SFTP client |
| [melbahja/goph](https://github.com/melbahja/goph) | SSH client |
| [stretchr/testify](https://github.com/stretchr/testify) | Testing toolkit |
| [Eun/go-hit](https://github.com/Eun/go-hit) | HTTP integration testing |
| [vmihailenco/taskq](https://github.com/vmihailenco/taskq) | Task queue |
| [getkin/kin-openapi](https://github.com/getkin/kin-openapi) | OpenAPI schema handling |
| [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) | S3 & cloud storage |

---

## Frontend Libraries

### UI & Components

| Library | Purpose |
|---------|---------|
| [shadcn/ui](https://ui.shadcn.com) | Component system and design patterns |
| [Radix UI](https://www.radix-ui.com) | Accessible, unstyled primitives |
| [Tailwind CSS](https://tailwindcss.com) | Utility-first CSS framework |
| [Lucide](https://lucide.dev) | Icon set |
| [cmdk](https://cmdk.paco.me) | Command palette component |
| [Sonner](https://sonner.emilkowal.ski) | Toast notifications |
| [Recharts](https://recharts.org) | Charting library |
| [Embla Carousel](https://www.embla-carousel.com) | Carousel component |
| [dnd-kit](https://dndkit.com) | Drag and drop |
| [react-resizable-panels](https://github.com/bvaughn/react-resizable-panels) | Resizable panel layouts |
| [React Hook Form](https://react-hook-form.com) | Form state management |
| [Zod](https://zod.dev) | Schema validation |
| [class-variance-authority](https://cva.style) | Variant-driven component styling |
| [tailwind-merge](https://github.com/dcastil/tailwind-merge) | Tailwind class merging |
| [clsx](https://github.com/lukeed/clsx) | Conditional class strings |
| [next-themes](https://github.com/pacocoursey/next-themes) | Theme switching |
| [Emotion](https://emotion.sh) | CSS-in-JS |
| [@reactour/tour](https://github.com/elrumordelaluz/reactour) | Guided product tours |
| [@stepperize/react](https://github.com/damianricobelli/stepperize) | Multi-step flows |

### State Management & Data

| Library | Purpose |
|---------|---------|
| [Redux Toolkit](https://redux-toolkit.js.org) | State management |
| [Redux Persist](https://github.com/rt2zz/redux-persist) | Persistent state |
| [date-fns](https://date-fns.org) | Date utilities |
| [uuid](https://github.com/uuidjs/uuid) | UUID generation |
| [yaml](https://github.com/eemeli/yaml) | YAML parsing |
| [nookies](https://github.com/maticzav/nookies) | Cookie helpers for Next.js |

### Terminal & Editor

| Library | Purpose |
|---------|---------|
| [xterm.js](https://xtermjs.org) | Terminal emulator |
| [React Ace](https://github.com/securingsincity/react-ace) | Code editor |
| [Streamdown](https://github.com/nicholasgriffintn/streamdown) | Streaming markdown renderer |
| [@xyflow/react](https://reactflow.dev) | Node-based graph UI |

### Auth & Analytics

| Library | Purpose |
|---------|---------|
| [Better Auth](https://www.better-auth.com) | Authentication framework |
| [PostHog](https://posthog.com) | Product analytics |
| [Mastra](https://mastra.ai) | AI agent client |

---

## Developer Tooling

| Tool | Purpose |
|------|---------|
| [ESLint](https://eslint.org) | JavaScript / TypeScript linting |
| [Prettier](https://prettier.io) | Code formatting |
| [Husky](https://typicode.github.io/husky) | Git hooks |
| [Commitlint](https://commitlint.js.org) | Commit message linting |
| [Air](https://github.com/air-verse/air) | Go live-reload for development |
| [SVGR](https://react-svgr.com) | SVG to React component transform |

---

## Documentation

| Tool | Purpose |
|------|---------|
| [VitePress](https://vitepress.dev) | Documentation site generator |
| [vitepress-openapi](https://github.com/enzonotario/vitepress-openapi) | OpenAPI docs integration |
| [Mermaid](https://mermaid.js.org) | Diagrams in markdown |

---

## Fonts

| Font | Designers | License |
|------|-----------|---------|
| [Roboto](https://fonts.google.com/specimen/Roboto) | Christian Robertson | Apache-2.0 |
| [JetBrains Mono](https://www.jetbrains.com/lp/mono/) | JetBrains | OFL-1.1 |
| [Inter](https://rsms.me/inter/) | Rasmus Andersson | OFL-1.1 |
| [DM Mono](https://fonts.google.com/specimen/DM+Mono) | Colophon Foundry | OFL-1.1 |

Fonts are served via [Google Fonts](https://fonts.google.com).

---

## Badges & Embeds

| Service | Usage |
|---------|-------|
| [Shields.io](https://shields.io) | README status badges |
| [contrib.rocks](https://contrib.rocks) | Contributor avatars grid |
| [Trendshift](https://trendshift.io) | Trending repository badge |

---

## README Design Reference

The README layout draws inspiration from the clean, badge-driven structure popularized by open-source projects such as [LobeChat](https://github.com/lobehub/lobe-chat).

---

## Cloud Providers (Referenced in Documentation)

Nixopus installation guides reference the following providers as compatible hosting options. Nixopus is not affiliated with or endorsed by any of them.

- [Hetzner](https://www.hetzner.com)
- [DigitalOcean](https://www.digitalocean.com)
- [Amazon Web Services](https://aws.amazon.com)
- [Google Cloud Platform](https://cloud.google.com)
- [Microsoft Azure](https://azure.microsoft.com)

---

## Container Base Images

The project uses the following public container images:

- `postgres:14-alpine`
- `redis:7-alpine`
- `caddy:2-alpine`
- `golang:1.25-alpine`
- `node:20-alpine` / `node:22-alpine`
- `alpine:3.18`

---

If we missed anyone or any project, please [open an issue](https://github.com/nixopus/nixopus/issues) or submit a pull request. Every contribution to the open-source ecosystem matters.

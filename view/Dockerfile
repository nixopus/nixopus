FROM node:18-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 600000 --production=false

FROM node:18-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ARG NEXT_PUBLIC_PORT=7443
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=$NEXT_PUBLIC_PORT
ENV NEXT_PUBLIC_PORT=$NEXT_PUBLIC_PORT

RUN yarn build

FROM node:18-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=$NEXT_PUBLIC_PORT
ENV NEXT_PUBLIC_PORT=$NEXT_PUBLIC_PORT

RUN addgroup -S nixopus && adduser -S nixopus -G nixopus

COPY --from=builder --chown=nixopus:nixopus /app/next.config.ts ./
COPY --from=builder --chown=nixopus:nixopus /app/public ./public
COPY --from=builder --chown=nixopus:nixopus /app/.next/standalone ./
COPY --from=builder --chown=nixopus:nixopus /app/.next/static ./.next/static

USER nixopus

EXPOSE ${NEXT_PUBLIC_PORT}

CMD ["sh", "-c", "node server.js"]

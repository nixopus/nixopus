FROM node:18-alpine AS builder

ARG NEXT_PUBLIC_BASE_URL=https://api.nixopus.com
ARG NEXT_PUBLIC_WEBSOCKET_URL=wss://api.nixopus.com

ENV NEXT_PUBLIC_BASE_URL=$NEXT_PUBLIC_BASE_URL
ENV NEXT_PUBLIC_WEBSOCKET_URL=$NEXT_PUBLIC_WEBSOCKET_URL
ENV NEXT_TELEMETRY_DISABLED=1

WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 600000

COPY . .

RUN yarn build

FROM node:18-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

RUN addgroup -g 1001 nixopus && \
    adduser -u 1001 -G nixopus -s /bin/sh -D nixopus

COPY --from=builder --chown=nixopus:nixopus /app/next.config.ts ./
COPY --from=builder --chown=nixopus:nixopus /app/public ./public
COPY --from=builder --chown=nixopus:nixopus /app/.next ./.next
COPY --from=builder --chown=nixopus:nixopus /app/node_modules ./node_modules
COPY --from=builder --chown=nixopus:nixopus /app/package.json ./
COPY --from=builder --chown=nixopus:nixopus /app/yarn.lock ./

RUN mkdir -p .next/cache && \
    chown -R nixopus:nixopus . && \
    chmod -R 755 .

USER nixopus

EXPOSE 3000

CMD ["yarn", "start"]
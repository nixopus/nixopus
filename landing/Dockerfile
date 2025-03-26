FROM node:18-alpine AS builder

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

RUN addgroup -g 1001 nixopus-landing && \
    adduser -u 1001 -G nixopus-landing -s /bin/sh -D nixopus-landing

COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/next.config.mjs ./
COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/public ./public
COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/.next ./.next
COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/node_modules ./node_modules
COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/package.json ./
COPY --from=builder --chown=nixopus-landing:nixopus-landing /app/yarn.lock ./

RUN mkdir -p .next/cache && \
    chown -R nixopus-landing:nixopus-landing . && \
    chmod -R 755 .

USER nixopus-landing

EXPOSE 3000

CMD ["yarn", "start"]
FROM node:18-alpine AS builder

ARG NEXT_PUBLIC_PORT=7443
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_PORT=$NEXT_PUBLIC_PORT

WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 600000

COPY . .

RUN yarn build && \
    yarn install --frozen-lockfile --network-timeout 600000 --production && \
    yarn cache clean

FROM node:18-alpine AS runner

ARG NEXT_PUBLIC_PORT=7443
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_PORT=$NEXT_PUBLIC_PORT
WORKDIR /app

RUN addgroup -S nixopus && adduser -S nixopus -G nixopus

COPY --from=builder --chown=nixopus:nixopus /app/next.config.ts ./
COPY --from=builder --chown=nixopus:nixopus /app/public ./public
COPY --from=builder --chown=nixopus:nixopus /app/.next ./.next
COPY --from=builder --chown=nixopus:nixopus /app/node_modules ./node_modules
COPY --from=builder --chown=nixopus:nixopus /app/package.json ./
COPY --from=builder --chown=nixopus:nixopus /app/yarn.lock ./

RUN mkdir -p .next/cache && \
    chown -R nixopus:nixopus . && \
    chmod -R 755 . && \
    rm -rf /app/node_modules/.cache

USER nixopus

EXPOSE ${NEXT_PUBLIC_PORT}

CMD yarn start -p ${NEXT_PUBLIC_PORT}
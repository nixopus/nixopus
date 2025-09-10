FROM golang:1.25-alpine AS builder
ARG API_PORT
RUN apk add --no-cache make git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ./migrations /app/migrations

RUN make build

FROM alpine:3.18
ARG API_PORT
RUN apk add --no-cache ca-certificates bash docker-cli

WORKDIR /app

COPY --from=builder /app/bin/nixopus-api .
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /etc/nixopus/ssh && \
    chmod 700 /etc/nixopus/ssh && \
    mkdir -p /etc/nixopus/docker-certs && \
    chmod 700 /etc/nixopus/docker-certs && \
    mkdir -p /etc/nixopus/configs && \
    chmod 700 /etc/nixopus/configs && \
    mkdir -p /etc/nixopus/db && \
    chmod 700 /etc/nixopus/db

EXPOSE ${API_PORT}

ENTRYPOINT ["./nixopus-api"]

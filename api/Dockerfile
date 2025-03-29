FROM alpine:3.18 as env-builder
ARG PORT
ARG SSH_HOST
ARG SSH_PORT
ARG SSH_USER
ARG SSH_PASSWORD
RUN echo "PORT=${PORT}" > .env
RUN echo "SSH_HOST=${SSH_HOST}" >> .env
RUN echo "SSH_PORT=${SSH_PORT}" >> .env
RUN echo "SSH_USER=${SSH_USER}" >> .env
RUN echo "SSH_PASSWORD=${SSH_PASSWORD}" >> .env

FROM golang:1.23.6-alpine AS builder
ARG PORT

RUN apk add --no-cache make git

WORKDIR /app

COPY --from=env-builder .env .

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ./migrations /app/migrations

RUN make build

FROM alpine:3.18

RUN apk add --no-cache ca-certificates bash

WORKDIR /app

COPY --from=builder /app/bin/nixopus-api .

COPY --from=builder /app/migrations ./migrations

COPY --from=env-builder .env .

EXPOSE ${PORT}

ENTRYPOINT ["./nixopus-api"]
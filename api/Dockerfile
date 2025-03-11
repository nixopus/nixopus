FROM alpine:3.18 as env-builder
ARG PORT
RUN echo "PORT=${PORT}" > .env

FROM golang:1.23.1-alpine AS builder
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

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/nixopus-api .

COPY --from=builder /app/migrations ./migrations

COPY --from=env-builder .env .

ENTRYPOINT ["./nixopus-api"]
FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk update && apk add --no-cache docker-cli

COPY memhog memhog

WORKDIR /build/memhog

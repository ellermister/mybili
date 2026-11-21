# syntax=docker/dockerfile:1

ARG NODE_VERSION=22
ARG GO_VERSION=1.24

FROM node:${NODE_VERSION}-bookworm-slim AS web

WORKDIR /src/web

RUN npm install -g pnpm@9.15.9

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

FROM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH
ARG APP_VERSION=2.0.0
RUN sed -i "s/^const Version = \".*\"$/const Version = \"${APP_VERSION}\"/" internal/version/version.go \
    && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/mybili ./cmd/mybili

FROM alpine:3.22

RUN apk add --no-cache ca-certificates ffmpeg tzdata

WORKDIR /app

COPY --from=build /out/mybili /usr/local/bin/mybili
COPY --from=web /src/web/dist /app/web/dist
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod 755 /usr/local/bin/docker-entrypoint.sh /usr/local/bin/mybili

ENV HTTP_ADDR=:80 \
    APP_TIMEZONE=Asia/Shanghai \
    TZ=Asia/Shanghai \
    APP_DEBUG=false \
    SQL_LOG_LEVEL=warn \
    DB_AUTO_MIGRATE=false \
    DB_DATABASE=/data/database.sqlite \
    DATA_DIR=/app/storage/app

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

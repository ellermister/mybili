ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-bullseye-slim AS build

WORKDIR /app

#ffmpeg - 支持多架构
ARG TARGETPLATFORM

RUN apt update && apt install -y xz-utils curl wget tar

# 根据目标平台下载对应的 ffmpeg
# 如果 TARGETPLATFORM 为空（非 buildx 环境），则使用 uname 检测当前架构
RUN PLATFORM="${TARGETPLATFORM:-linux/$(uname -m)}" && \
    case "${PLATFORM}" in \
        "linux/amd64"|"linux/x86_64") \
            FFMPEG_URL="https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-6.0.1-amd64-static.tar.xz" \
            ;; \
        "linux/arm64"|"linux/aarch64") \
            FFMPEG_URL="https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-6.0.1-arm64-static.tar.xz" \
            ;; \
        *) \
            echo "Unsupported platform: ${PLATFORM}" && exit 1 \
            ;; \
    esac && \
    curl -L ${FFMPEG_URL} -o /tmp/ffmpeg.tar.xz && \
    cd /tmp && \
    xz -d /tmp/ffmpeg.tar.xz && \
    tar -xf ffmpeg.tar -C /tmp --strip-components=1 && \
    cp /tmp/ffmpeg /usr/local/bin/ffmpeg && \
    chmod +x /usr/local/bin/ffmpeg && \
    cp /tmp/ffprobe /usr/local/bin/ffprobe && \
    chmod +x /usr/local/bin/ffprobe

# 根据目标平台下载对应的 yt-dlp
# 如果 TARGETPLATFORM 为空（非 buildx 环境），则使用 uname 检测当前架构
RUN PLATFORM="${TARGETPLATFORM:-linux/$(uname -m)}" && \
    case "${PLATFORM}" in \
        "linux/amd64"|"linux/x86_64") \
            YT_DLP_ARCH="yt-dlp_linux" \
            ;; \
        "linux/arm64"|"linux/aarch64") \
            YT_DLP_ARCH="yt-dlp_linux_aarch64" \
            ;; \
        *) \
            echo "Unsupported platform: ${PLATFORM}" && exit 1 \
            ;; \
    esac && \
    wget -O /usr/local/bin/yt-dlp_linux \
        "https://github.com/yt-dlp/yt-dlp/releases/download/2026.02.21/${YT_DLP_ARCH}" && \
    chmod 775 /usr/local/bin/yt-dlp_linux && \
    chown root:root /usr/local/bin/yt-dlp_linux

# 根据目标平台下载对应的 frankenphp
# 如果 TARGETPLATFORM 为空（非 buildx 环境），则使用 uname 检测当前架构
RUN PLATFORM="${TARGETPLATFORM:-linux/$(uname -m)}" && \
    case "${PLATFORM}" in \
        "linux/amd64"|"linux/x86_64") \
            FRANKENPHP_ARCH="frankenphp-linux-x86_64" \
            ;; \
        "linux/arm64"|"linux/aarch64") \
            FRANKENPHP_ARCH="frankenphp-linux-aarch64" \
            ;; \
        *) \
            echo "Unsupported platform: ${PLATFORM}" && exit 1 \
            ;; \
    esac && \
    wget -O /usr/local/bin/frankenphp \
        "https://github.com/dunglas/frankenphp/releases/download/v1.2.5/${FRANKENPHP_ARCH}" && \
    chmod 775 /usr/local/bin/frankenphp && \
    chown root:root /usr/local/bin/frankenphp

# 根据目标平台下载对应的 supervisord
# 如果 TARGETPLATFORM 为空（非 buildx 环境），则使用 uname 检测当前架构
RUN PLATFORM="${TARGETPLATFORM:-linux/$(uname -m)}" && \
    case "${PLATFORM}" in \
        "linux/amd64"|"linux/x86_64") \
            SUPERVISORD_URL="https://github.com/ochinchina/supervisord/releases/download/v0.7.3/supervisord_0.7.3_Linux_64-bit.tar.gz" \
            ;; \
        "linux/arm64"|"linux/aarch64") \
            SUPERVISORD_URL="https://github.com/ochinchina/supervisord/releases/download/v0.7.3/supervisord_0.7.3_Linux_ARM64.tar.gz" \
            ;; \
        *) \
            echo "Unsupported platform: ${PLATFORM}" && exit 1 \
            ;; \
    esac && \
    curl -L ${SUPERVISORD_URL} -o /tmp/supervisord.tar.gz && \
    tar -xzf /tmp/supervisord.tar.gz -C /usr/local/bin/ --strip-components=1 && \
    chmod +x /usr/local/bin/supervisord

COPY . .

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

FROM phpswoole/swoole:php8.3-alpine

# 重新声明 ARG，确保能从构建参数接收版本号
ARG APP_VERSION=1.0.0
ARG WEBSITE_ID
ARG TARGETPLATFORM

WORKDIR /app

ENV PHPRC=/etc/php.ini

ENV COMPOSER_ALLOW_SUPERUSER=1

COPY --from=composer/composer:2-bin /composer /usr/bin/composer

COPY --from=build /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=build /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=build /usr/local/bin/yt-dlp_linux /usr/local/bin/yt-dlp_linux
COPY --from=build /usr/local/bin/frankenphp /usr/local/bin/frankenphp
COPY --from=build /usr/local/bin/supervisord /usr/local/bin/supervisord

RUN docker-php-ext-install pcntl

COPY . .

COPY  ./deploy/files/ /
COPY --from=build /app/public/ /app/public/


RUN composer install \
    --ignore-platform-reqs \
    --classmap-authoritative \
    --no-interaction \
    --no-ansi \
    --no-dev \
    && rm -rf /root/.composer
    

RUN cp .env.example .env \
    && php artisan key:generate \
    && rm -f public/storage && php artisan storage:link \
    && php artisan migrate --force \
    && php artisan octane:install --server=frankenphp


ENV APP_VERSION=${APP_VERSION}
ENV WEBSITE_ID=${WEBSITE_ID}
ENV DB_DATABASE=/data/database.sqlite

CMD ["/usr/local/bin/supervisord", "-c", "/etc/supervisord.conf"]

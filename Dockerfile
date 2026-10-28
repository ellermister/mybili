ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-bookworm-slim AS build

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
            YT_DLP_ARCH="yt-dlp_musllinux" \
            ;; \
        "linux/arm64"|"linux/aarch64") \
            YT_DLP_ARCH="yt-dlp_musllinux_aarch64" \
            ;; \
        *) \
            echo "Unsupported platform: ${PLATFORM}" && exit 1 \
            ;; \
    esac && \
    wget -O /usr/local/bin/yt-dlp_linux \
        "https://github.com/yt-dlp/yt-dlp/releases/latest/download/${YT_DLP_ARCH}" && \
    chmod 775 /usr/local/bin/yt-dlp_linux && \
    chown root:root /usr/local/bin/yt-dlp_linux

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

RUN npm install -g pnpm@latest-11
RUN pnpm install
RUN pnpm build

# 网页由 FrankenPHP 处理，不能只升 Swoole 镜像里的 CLI PHP。
# 官方静态二进制从 1.13.2 起只带 PHP 8.5；这份镜像把 FrankenPHP 1.13.4 链到 PHP 8.3.36。
FROM dunglas/frankenphp:1.13.4-php8.3.36-alpine

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
COPY --from=build /usr/local/bin/supervisord /usr/local/bin/supervisord

# pcntl：队列和定时任务；redis：Horizon（phpredis）；gd：封面主色；pdo_mysql：可选 MySQL 连接
RUN install-php-extensions pcntl redis gd pdo_mysql

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

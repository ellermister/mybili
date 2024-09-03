ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-bullseye-slim as build

WORKDIR /app

ENV PHPRC=/etc/php.ini
ENV PHP_BINARY=/usr/local/bin/php

ADD --chown=root:root --chmod=775 \
    https://github.com/dunglas/frankenphp/releases/download/v1.2.5/frankenphp-linux-x86_64 /usr/local/bin/frankenphp

ENV COMPOSER_ALLOW_SUPERUSER=1
COPY --from=composer/composer:2-bin /composer /usr/bin/composer

ADD https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-6.0.1-amd64-static.tar.xz /tmp

RUN apt update \
    && apt install -y xz-utils  \
    && cd /tmp \
    && xz -d /tmp/ffmpeg-6.0.1* \
    && tar -xf ffmpeg-6.0.1* -C /tmp --strip-components=1 \
    && cp /tmp/ffmpeg /usr/local/bin/ffmpeg \
    && chmod +x /usr/local/bin/ffmpeg

COPY . .

COPY  ./deploy/files/ /

RUN  chmod +x /usr/local/bin/php

COPY --from=ochinchina/supervisord:latest /usr/local/bin/supervisord /usr/local/bin/supervisord


RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

RUN composer install \
    --ignore-platform-reqs \
    --classmap-authoritative \
    --no-interaction \
    --no-ansi \
    --no-dev 

# Run
FROM debian:12

WORKDIR /app

ENV PHPRC=/etc/php.ini
ENV PHP_BINARY=/usr/local/bin/php

ADD --chown=root:root --chmod=775 \
    https://github.com/dunglas/frankenphp/releases/download/v1.2.5/frankenphp-linux-x86_64 /usr/local/bin/frankenphp

ADD --chown=root:root --chmod=775 \
https://github.com/yt-dlp/yt-dlp/releases/download/2024.08.06/yt-dlp_linux /usr/local/bin/yt-dlp_linux

COPY  ./deploy/files/ /

COPY . .
COPY --from=build /app/vendor/ /app/vendor/
COPY --from=build /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg

RUN  chmod +x /usr/local/bin/php
RUN rm -f public/storage && php artisan storage:link

CMD ["/usr/local/bin/supervisord", "-c", "/etc/supervisord.conf"]

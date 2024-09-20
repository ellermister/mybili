ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-bullseye-slim as build

WORKDIR /app

COPY . .

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

#ffmpeg
ADD https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-6.0.1-amd64-static.tar.xz /tmp

RUN apt update \
    && apt install -y xz-utils  \
    && cd /tmp \
    && xz -d /tmp/ffmpeg-6.0.1* \
    && tar -xf ffmpeg-6.0.1* -C /tmp --strip-components=1 \
    && cp /tmp/ffmpeg /usr/local/bin/ffmpeg \
    && chmod +x /usr/local/bin/ffmpeg


FROM phpswoole/swoole:php8.3-alpine

WORKDIR /app

ENV PHPRC=/etc/php.ini

ADD --chown=root:root --chmod=775 \
https://github.com/yt-dlp/yt-dlp/releases/download/2024.08.06/yt-dlp_linux /usr/local/bin/yt-dlp_linux

ADD --chown=root:root --chmod=775 \
    https://github.com/dunglas/frankenphp/releases/download/v1.2.5/frankenphp-linux-x86_64 /usr/local/bin/frankenphp

ENV COMPOSER_ALLOW_SUPERUSER=1

COPY --from=composer/composer:2-bin /composer /usr/bin/composer

COPY --from=build /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg

COPY --from=ochinchina/supervisord:latest /usr/local/bin/supervisord /usr/local/bin/supervisord

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
    && php artisan migrate

CMD ["/usr/local/bin/supervisord", "-c", "/etc/supervisord.conf"]

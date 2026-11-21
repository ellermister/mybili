#!/bin/sh
set -eu

mkdir -p \
    /data \
    /app/storage/app/public/videos \
    /app/storage/app/public/images \
    /app/storage/app/public/medias

/usr/local/bin/mybili migrate

# 迁移只在启动时跑一次，避免 server 和 worker 同时改表。
export DB_AUTO_MIGRATE=false

/usr/local/bin/mybili worker &
worker=$!
/usr/local/bin/mybili server &
server=$!

stop() {
    kill -TERM "$server" "$worker" 2>/dev/null || true
    wait "$server" 2>/dev/null || true
    wait "$worker" 2>/dev/null || true
}

trap 'stop; exit 0' TERM INT

while :; do
    if ! kill -0 "$server" 2>/dev/null; then
        status=0
        wait "$server" || status=$?
        kill -TERM "$worker" 2>/dev/null || true
        wait "$worker" 2>/dev/null || true
        exit "$status"
    fi
    if ! kill -0 "$worker" 2>/dev/null; then
        status=0
        wait "$worker" || status=$?
        kill -TERM "$server" 2>/dev/null || true
        wait "$server" 2>/dev/null || true
        exit "$status"
    fi
    sleep 1
done

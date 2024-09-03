#!/bin/bash
ls -ahl /tmp/blog.img
docker load < /tmp/blog.img

cd /root/blog
git reset --hard HEAD
git pull origin main

docker-compose up -d
docker image prune -a -f
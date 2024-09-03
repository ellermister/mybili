#!/bin/sh
echo "down -> $1 > $2";
echo $(pwd);
yt-dlp_linux -f "bestvideo+bestaudio/best" \
    --no-playlist \
    --cookies=./storage/app/cookie.txt \
    -o "$2" \
    $1
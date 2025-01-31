#!/bin/sh
# $1: URL
# $2: 输出路径
# $3: 可选，指定视频序号（如 1 或 1,2,3 或 1-3）

echo "down -> $1 > $2";
echo $(pwd);

if [ -n "$3" ]; then
    # 如果提供了第三个参数，使用 --playlist-items
    yt-dlp_linux -f "bestvideo+bestaudio/best" \
        --playlist-items "$3" \
        --cookies=./storage/app/cookie.txt \
        -o "$2" \
        $1
else
    # 保持原有的单视频下载逻辑
    yt-dlp_linux -f "bestvideo+bestaudio/best" \
        --no-playlist \
        --cookies=./storage/app/cookie.txt \
        -o "$2" \
        $1
fi
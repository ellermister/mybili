-- 收藏夹曾经把剧集 ep id 写成 videos.id。
-- 正确的 aid 行已经有分 P。这条 SQL 只删同一 bvid 下没有分 P 的那条非音频空壳。
-- 音频 type=12 即使和某条视频撞号也不删。有分 P 的行、下载文件都不动。
-- 先跑下面的 SELECT 看清要删的行，确认后再跑 DELETE。

SELECT
    stub.id AS stub_id,
    stub.type AS stub_type,
    stub.title AS stub_title,
    stub.bvid,
    real.id AS real_id,
    real.title AS real_title
FROM videos AS stub
JOIN videos AS real
    ON real.bvid = stub.bvid
    AND real.id <> stub.id
    AND real.type <> 12
WHERE stub.type <> 12
  AND stub.bvid <> ''
  AND NOT EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = stub.id)
  AND EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = real.id);

BEGIN;

DROP TABLE IF EXISTS fav_stub_ids;
CREATE TEMP TABLE fav_stub_ids AS
SELECT stub.id
FROM videos AS stub
WHERE stub.type <> 12
  AND stub.bvid <> ''
  AND NOT EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = stub.id)
  AND EXISTS (
      SELECT 1
      FROM videos AS real
      WHERE real.bvid = stub.bvid
        AND real.id <> stub.id
        AND real.type <> 12
        AND EXISTS (SELECT 1 FROM video_parts WHERE video_parts.video_id = real.id)
  );

DELETE FROM favorite_list_videos
WHERE video_id IN (SELECT id FROM fav_stub_ids);

DELETE FROM subscription_videos
WHERE video_id IN (SELECT id FROM fav_stub_ids);

DELETE FROM coverables
WHERE coverable_type = 'App\Models\Video'
  AND coverable_id IN (SELECT id FROM fav_stub_ids);

DELETE FROM danmaku
WHERE video_id IN (SELECT id FROM fav_stub_ids);

DELETE FROM audio_parts
WHERE video_id IN (SELECT id FROM fav_stub_ids);

DELETE FROM videos
WHERE id IN (SELECT id FROM fav_stub_ids);

COMMIT;

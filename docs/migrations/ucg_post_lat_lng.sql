-- UCG 帖子可选坐标（WGS84），供 Feed GEO 索引与距离展示。
-- 在 UCG_DB_LINK 对应库、部署本 change 前执行。
ALTER TABLE ucg_post
  ADD COLUMN lat DOUBLE NULL COMMENT '发帖纬度 WGS84' AFTER ip_location,
  ADD COLUMN lng DOUBLE NULL COMMENT '发帖经度 WGS84' AFTER lat;

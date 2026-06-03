-- device 库 user 表：记录设备最近一次带 deviceNo 的对外 HTTP API（由网关 touch 写入）
ALTER TABLE `user`
  ADD COLUMN `last_api_path` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '最近 HTTP 接口 METHOD /path',
  ADD COLUMN `last_api_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近 HTTP 接口时间 Unix 秒';

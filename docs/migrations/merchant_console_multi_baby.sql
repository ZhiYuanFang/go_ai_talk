-- =============================================================================
-- 商户后台 + 一账号多宝宝（手工参考；应用进程启动 Ensure 幂等建表/补列）
-- device 库：生产 ai_voice_device / 测试 ai_voice_device_test → device-service
-- cash 库：生产 ai_voice_cash / 测试 ai_voice_cash_test → cash-service
--
-- 执行前请备份。重复执行依赖 IF NOT EXISTS / 忽略 Duplicate column。
-- =============================================================================

-- —— device 库 ——
ALTER TABLE wx
  ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'parent' COMMENT 'parent|staff 账号级单值' AFTER password;

CREATE TABLE IF NOT EXISTS merchant (
  id            BIGINT       NOT NULL AUTO_INCREMENT,
  store_name    VARCHAR(64)  NOT NULL COMMENT '店名，全局唯一',
  password_hash VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'bcrypt',
  created_at    BIGINT       NOT NULL DEFAULT 0,
  updated_at    BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_name (store_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='B 端商户主账号';

CREATE TABLE IF NOT EXISTS wx_baby_member (
  id         BIGINT      NOT NULL AUTO_INCREMENT,
  wx_id      BIGINT      NOT NULL,
  device_no  VARCHAR(64) NOT NULL,
  source     VARCHAR(32) NOT NULL DEFAULT 'self' COMMENT 'self|merchant_assign|bind',
  created_at BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wx_device (wx_id, device_no),
  KEY idx_device (device_no),
  KEY idx_wx (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账号↔宝宝成员真源';

CREATE TABLE IF NOT EXISTS merchant_baby (
  id           BIGINT      NOT NULL AUTO_INCREMENT,
  merchant_id  BIGINT      NOT NULL,
  device_no    VARCHAR(64) NOT NULL,
  created_at   BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_device (device_no),
  KEY idx_merchant (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户拥有的虚拟宝宝';

-- 回填示例（EnsureBackfillWxBabyMembers 启动时执行）：
-- INSERT IGNORE INTO wx_baby_member (wx_id, device_no, source, created_at)
-- SELECT id, device_no, 'self', UNIX_TIMESTAMP() FROM wx WHERE device_no IS NOT NULL AND device_no <> '';

-- —— cash 库 ——
CREATE TABLE IF NOT EXISTS merchant_product (
  product_code       VARCHAR(64)  NOT NULL,
  title              VARCHAR(128) NOT NULL DEFAULT '',
  price_fen          INT          NOT NULL DEFAULT 100000 COMMENT '默认 1000 元/月',
  period_days        INT          NOT NULL DEFAULT 30,
  status             TINYINT      NOT NULL DEFAULT 1,
  updated_at         BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (product_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户连续订阅商品';

CREATE TABLE IF NOT EXISTS merchant_entitlement (
  merchant_id  BIGINT       NOT NULL,
  expire_at    BIGINT       NOT NULL DEFAULT 0 COMMENT '权威到期 Unix 秒',
  source       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT 'hub_grant|alipay_period',
  updated_at   BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户权益（Hub 与支付宝共用 expire_at）';

CREATE TABLE IF NOT EXISTS merchant_order (
  id              BIGINT       NOT NULL AUTO_INCREMENT,
  order_no        VARCHAR(64)  NOT NULL,
  merchant_id     BIGINT       NOT NULL,
  product_code    VARCHAR(64)  NOT NULL,
  channel         VARCHAR(32)  NOT NULL DEFAULT 'alipay',
  amount_fen      INT          NOT NULL DEFAULT 0,
  status          VARCHAR(16)  NOT NULL DEFAULT 'created',
  channel_txn_id  VARCHAR(128) NOT NULL DEFAULT '',
  agreement_no    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '支付宝签约号等',
  created_at      BIGINT       NOT NULL DEFAULT 0,
  paid_at         BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_merchant (merchant_id),
  KEY idx_channel_txn (channel, channel_txn_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户订阅订单';

INSERT IGNORE INTO merchant_product (product_code, title, price_fen, period_days, status, updated_at)
VALUES ('merchant_sub_monthly', '商户连续订阅（月）', 100000, 30, 1, UNIX_TIMESTAMP());

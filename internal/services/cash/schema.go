package cash

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	// ProductMonthly19 一期唯一 SKU：19 元 / 30 天。
	ProductMonthly19 = "vip_monthly_19"
	ProductPriceFen  = 1900
	// ProductOriginalPriceFen 种子划线原价（分）；0 表示不展示。运维可用 SQL 改库。
	ProductOriginalPriceFen = 9900
	ProductDurationD        = 30

	ChannelAlipay   = "alipay"
	ChannelAppleIAP = "apple_iap"

	OrderCreated = "created"
	OrderPaid    = "paid"
	OrderFailed  = "failed"
	OrderClosed  = "closed"
)

// EnsureSchema 创建 VIP 表、商业功能开通表并种子一期商品（幂等）。
func EnsureSchema(ctx context.Context) error {
	db := g.DB()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS vip_product (
  product_code         VARCHAR(64)  NOT NULL,
  title                VARCHAR(128) NOT NULL DEFAULT '',
  price_fen            INT          NOT NULL DEFAULT 0,
  original_price_fen   INT          NOT NULL DEFAULT 0,
  duration_days        INT          NOT NULL DEFAULT 30,
  apple_product_id     VARCHAR(128) NOT NULL DEFAULT '',
  status               TINYINT      NOT NULL DEFAULT 1,
  updated_at           BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (product_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS vip_order (
  id              BIGINT       NOT NULL AUTO_INCREMENT,
  order_no        VARCHAR(64)  NOT NULL,
  wx_id           BIGINT       NOT NULL,
  product_code    VARCHAR(64)  NOT NULL,
  channel         VARCHAR(32)  NOT NULL,
  amount_fen      INT          NOT NULL DEFAULT 0,
  currency        VARCHAR(8)   NOT NULL DEFAULT 'CNY',
  status          VARCHAR(16)  NOT NULL DEFAULT 'created',
  channel_txn_id  VARCHAR(128) NOT NULL DEFAULT '',
  created_at      BIGINT       NOT NULL DEFAULT 0,
  paid_at         BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_wx_created (wx_id, created_at),
  KEY idx_channel_txn (channel, channel_txn_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS vip_entitlement (
  wx_id       BIGINT NOT NULL,
  expire_at   BIGINT NOT NULL DEFAULT 0,
  updated_at  BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		// —— 商业功能开通域（commercial-feature-entitlement）——
		`CREATE TABLE IF NOT EXISTS feature_def (
  feature_id             VARCHAR(64)  NOT NULL,
  title                  VARCHAR(128) NOT NULL DEFAULT '',
  description            VARCHAR(512) NOT NULL DEFAULT '',
  unlock_methods         VARCHAR(128) NOT NULL DEFAULT '',
  duration_days          INT          NOT NULL DEFAULT 0,
  default_allowed_count  INT          NOT NULL DEFAULT 0,
  status                 TINYINT      NOT NULL DEFAULT 1,
  sort_order             INT          NOT NULL DEFAULT 0,
  updated_at             BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_product (
  product_code         VARCHAR(64)  NOT NULL,
  feature_id           VARCHAR(64)  NOT NULL,
  grant_kind           VARCHAR(32)  NOT NULL DEFAULT 'entitlement',
  grant_quantity       INT          NOT NULL DEFAULT 1,
  price_fen            INT          NOT NULL DEFAULT 0,
  original_price_fen   INT          NOT NULL DEFAULT 0,
  duration_days        INT          NOT NULL DEFAULT 0,
  apple_product_id     VARCHAR(128) NOT NULL DEFAULT '',
  status               TINYINT      NOT NULL DEFAULT 1,
  updated_at           BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (product_code),
  KEY idx_feature (feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_invite_code (
  code                 VARCHAR(64)  NOT NULL,
  owner_wx_id          BIGINT       NOT NULL,
  expires_at           BIGINT       NOT NULL DEFAULT 0,
  max_redemptions      INT          NOT NULL DEFAULT 0,
  redeemed_count       INT          NOT NULL DEFAULT 0,
  grant_duration_days  INT          NOT NULL DEFAULT 0,
  status               TINYINT      NOT NULL DEFAULT 1,
  created_at           BIGINT       NOT NULL DEFAULT 0,
  updated_at           BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (code),
  UNIQUE KEY uk_owner_wx (owner_wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_invite_code_feature (
  code            VARCHAR(64) NOT NULL,
  feature_id      VARCHAR(64) NOT NULL,
  grant_quantity  INT         NOT NULL DEFAULT 1,
  PRIMARY KEY (code, feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_invite_redeemer_bind (
  redeemer_wx_id  BIGINT NOT NULL,
  owner_wx_id     BIGINT NOT NULL,
  bound_at        BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (redeemer_wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_invite_feature_grant (
  redeemer_wx_id  BIGINT      NOT NULL,
  feature_id      VARCHAR(64) NOT NULL,
  code            VARCHAR(64) NOT NULL DEFAULT '',
  device_no       VARCHAR(64) NOT NULL DEFAULT '',
  redeemed_at     BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (redeemer_wx_id, feature_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_invite_redemption (
  id               BIGINT      NOT NULL AUTO_INCREMENT,
  code             VARCHAR(64) NOT NULL,
  owner_wx_id      BIGINT      NOT NULL,
  redeemer_wx_id   BIGINT      NOT NULL,
  device_no        VARCHAR(64) NOT NULL DEFAULT '',
  feature_id       VARCHAR(64) NOT NULL,
  redeemed_at      BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_code_time (code, redeemed_at),
  KEY idx_redeemer (redeemer_wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_entitlement (
  id              BIGINT      NOT NULL AUTO_INCREMENT,
  device_no       VARCHAR(64) NOT NULL,
  feature_id      VARCHAR(64) NOT NULL,
  unlock_method   VARCHAR(32) NOT NULL DEFAULT '',
  expires_at      BIGINT      NOT NULL DEFAULT 0,
  quantity        INT         NOT NULL DEFAULT 0,
  source_ref      VARCHAR(128) NOT NULL DEFAULT '',
  created_at      BIGINT      NOT NULL DEFAULT 0,
  updated_at      BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_device_feature (device_no, feature_id),
  KEY idx_device (device_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_allowed_count (
  device_no                VARCHAR(64) NOT NULL,
  allowed_count            INT         NOT NULL DEFAULT 0,
  full_access              TINYINT     NOT NULL DEFAULT 0,
  full_access_expires_at   BIGINT      NOT NULL DEFAULT 0,
  updated_at               BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (device_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS feature_order (
  id              BIGINT       NOT NULL AUTO_INCREMENT,
  order_no        VARCHAR(64)  NOT NULL,
  device_no       VARCHAR(64)  NOT NULL,
  wx_id           BIGINT       NOT NULL DEFAULT 0,
  product_code    VARCHAR(64)  NOT NULL,
  channel         VARCHAR(32)  NOT NULL,
  amount_fen      INT          NOT NULL DEFAULT 0,
  currency        VARCHAR(8)   NOT NULL DEFAULT 'CNY',
  status          VARCHAR(16)  NOT NULL DEFAULT 'created',
  channel_txn_id  VARCHAR(128) NOT NULL DEFAULT '',
  created_at      BIGINT       NOT NULL DEFAULT 0,
  paid_at         BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_device_created (device_no, created_at),
  KEY idx_channel_txn (channel, channel_txn_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, sql := range stmts {
		if _, err := db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	// 已有库升级：补列（重复执行忽略 Duplicate column）。
	alterCols := []string{
		`ALTER TABLE vip_product ADD COLUMN original_price_fen INT NOT NULL DEFAULT 0`,
		`ALTER TABLE feature_def ADD COLUMN default_allowed_count INT NOT NULL DEFAULT 0`,
		`ALTER TABLE feature_allowed_count ADD COLUMN full_access TINYINT NOT NULL DEFAULT 0`,
		`ALTER TABLE feature_allowed_count ADD COLUMN full_access_expires_at BIGINT NOT NULL DEFAULT 0`,
	}
	for _, alterSQL := range alterCols {
		if _, err := db.Exec(ctx, alterSQL); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
				return err
			}
		}
	}
	applePID := strings.TrimSpace(os.Getenv("CASH_APPLE_PRODUCT_ID"))
	if applePID == "" {
		if v, err := g.Cfg().Get(ctx, "cash.appleProductId"); err == nil && v != nil {
			applePID = strings.TrimSpace(v.String())
		}
	}
	now := time.Now().Unix()
	// 不覆盖已有 price_fen / original_price_fen，便于运维手工 SQL 改价后重启仍保留。
	_, err := db.Exec(ctx, `
INSERT INTO vip_product (product_code, title, price_fen, original_price_fen, duration_days, apple_product_id, status, updated_at)
VALUES (?, 'VIP月会员', ?, ?, ?, ?, 1, ?)
ON DUPLICATE KEY UPDATE
  title=VALUES(title),
  duration_days=VALUES(duration_days),
  apple_product_id=IF(VALUES(apple_product_id)='', apple_product_id, VALUES(apple_product_id)),
  original_price_fen=IF(original_price_fen=0, VALUES(original_price_fen), original_price_fen),
  status=1,
  updated_at=VALUES(updated_at)`,
		ProductMonthly19, ProductPriceFen, ProductOriginalPriceFen, ProductDurationD, applePID, now)
	if err != nil {
		return err
	}
	// 种子预测开通功能定义（可停用；运营可用 Admin 改文案/默认条数；不覆盖已有 default_allowed_count）。
	_, err = db.Exec(ctx, `
INSERT INTO feature_def (feature_id, title, description, unlock_methods, duration_days, default_allowed_count, status, sort_order, updated_at)
VALUES (?, '预测事项开通数量', '增加可展示的预测事项数量', 'payment,invite_code,ad', 0, 0, 1, 10, ?)
ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
		FeatureIDPredictionUnlock, now)
	return err
}

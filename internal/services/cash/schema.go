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

// EnsureSchema 创建 VIP 表并种子一期商品（幂等）。
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
	}
	for _, sql := range stmts {
		if _, err := db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	// 已有库升级：补原价列（重复执行忽略 Duplicate column）。
	if _, err := db.Exec(ctx, `ALTER TABLE vip_product ADD COLUMN original_price_fen INT NOT NULL DEFAULT 0`); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
			return err
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
	return err
}

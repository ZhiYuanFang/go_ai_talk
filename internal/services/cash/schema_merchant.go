package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// MerchantProductCodeMonthly 商户连续订阅月费商品编码。
	MerchantProductCodeMonthly = "merchant_sub_monthly"
	// MerchantDefaultPriceFen 默认 1000 元/月。
	MerchantDefaultPriceFen = 100000
	// MerchantDefaultPeriodDays 默认计费周期天数。
	MerchantDefaultPeriodDays = 30

	MerchantEntitlementSourceHub    = "hub_grant"
	MerchantEntitlementSourceAlipay = "alipay_period"
)

// EnsureMerchantSchema 幂等创建商户权益/商品/订单表并种子默认月费商品（cash 库）。
func EnsureMerchantSchema(ctx context.Context) error {
	db := g.DB()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS merchant_product (
  product_code       VARCHAR(64)  NOT NULL,
  title              VARCHAR(128) NOT NULL DEFAULT '',
  price_fen          INT          NOT NULL DEFAULT 100000 COMMENT '默认 1000 元/月',
  period_days        INT          NOT NULL DEFAULT 30,
  status             TINYINT      NOT NULL DEFAULT 1,
  updated_at         BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (product_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户连续订阅商品'`,
		`CREATE TABLE IF NOT EXISTS merchant_entitlement (
  merchant_id  BIGINT       NOT NULL,
  expire_at    BIGINT       NOT NULL DEFAULT 0 COMMENT '权威到期 Unix 秒',
  source       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT 'hub_grant|alipay_period',
  updated_at   BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户权益（Hub 与支付宝共用 expire_at）'`,
		`CREATE TABLE IF NOT EXISTS merchant_order (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户订阅订单'`,
	}
	for _, sql := range stmts {
		if _, err := db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	now := time.Now().Unix()
	_, err := db.Exec(ctx, `
INSERT IGNORE INTO merchant_product (product_code, title, price_fen, period_days, status, updated_at)
VALUES (?, '商户连续订阅（月）', ?, ?, 1, ?)`,
		MerchantProductCodeMonthly, MerchantDefaultPriceFen, MerchantDefaultPeriodDays, now)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[cash-schema] merchant tables ensured; product=%s price_fen=%d",
		MerchantProductCodeMonthly, MerchantDefaultPriceFen)
	return nil
}

// IsMerchantEntitled 商户权益是否有效（expire_at > now）。
func IsMerchantEntitled(ctx context.Context, merchantID int64) (bool, int64, error) {
	if merchantID <= 0 {
		return false, 0, nil
	}
	one, err := g.DB().Model("merchant_entitlement").Ctx(ctx).
		Where("merchant_id", merchantID).Limit(1).One()
	if err != nil {
		return false, 0, err
	}
	if one.IsEmpty() {
		return false, 0, nil
	}
	expireAt := one["expire_at"].Int64()
	return expireAt > time.Now().Unix(), expireAt, nil
}

// SetMerchantExpireAtAbsolute Hub 绝对授期；expireAt 必须严格大于 now。
func SetMerchantExpireAtAbsolute(ctx context.Context, merchantID, expireAt int64, source string) error {
	if merchantID <= 0 {
		return errMerchantIDInvalid
	}
	now := time.Now().Unix()
	if expireAt <= now {
		return errMerchantExpireAtInvalid
	}
	src := strings.TrimSpace(source)
	if src == "" {
		src = MerchantEntitlementSourceHub
	}
	_, err := g.DB().Model("merchant_entitlement").Ctx(ctx).Data(g.Map{
		"merchant_id": merchantID,
		"expire_at":   expireAt,
		"source":      src,
		"updated_at":  now,
	}).OnDuplicate("expire_at", "source", "updated_at").Save()
	return err
}

// ExtendMerchantExpireAt 按周期延长 expire_at = max(now, current) + periodDays*86400。
func ExtendMerchantExpireAt(ctx context.Context, merchantID int64, periodDays int, source string) (int64, error) {
	if merchantID <= 0 {
		return 0, errMerchantIDInvalid
	}
	if periodDays < 1 {
		periodDays = MerchantDefaultPeriodDays
	}
	now := time.Now().Unix()
	_, current, err := IsMerchantEntitled(ctx, merchantID)
	if err != nil {
		return 0, err
	}
	base := now
	if current > base {
		base = current
	}
	newExpire := base + int64(periodDays)*86400
	src := strings.TrimSpace(source)
	if src == "" {
		src = MerchantEntitlementSourceAlipay
	}
	_, err = g.DB().Model("merchant_entitlement").Ctx(ctx).Data(g.Map{
		"merchant_id": merchantID,
		"expire_at":   newExpire,
		"source":      src,
		"updated_at":  now,
	}).OnDuplicate("expire_at", "source", "updated_at").Save()
	return newExpire, err
}

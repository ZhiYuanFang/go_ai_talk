package cash

import (
	"context"
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
	// ChannelAdmin Hub 手工授 VIP/功能（0 元 paid 订单 + grant_reason）。
	ChannelAdmin = "admin"

	OrderCreated  = "created"
	OrderPaid     = "paid"
	OrderRefunded = "refunded" // Apple ASN REFUND 等退款语义
	// OrderRevoked Hub 撤销手工授：权益立即过期，不用 refunded（那是支付退款）。
	OrderRevoked = "revoked"
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
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  order_no           VARCHAR(64)  NOT NULL,
  wx_id              BIGINT       NOT NULL,
  product_code       VARCHAR(64)  NOT NULL,
  channel            VARCHAR(32)  NOT NULL,
  amount_fen         INT          NOT NULL DEFAULT 0,
  currency           VARCHAR(8)   NOT NULL DEFAULT 'CNY',
  status             VARCHAR(16)  NOT NULL DEFAULT 'created',
  channel_txn_id     VARCHAR(128) NOT NULL DEFAULT '',
  app_account_token  CHAR(36)     NULL DEFAULT NULL COMMENT 'Apple StoreKit appAccountToken(UUID)；仅 apple_iap',
  grant_reason       VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Admin 手工授理由；支付单为空',
  created_at         BIGINT       NOT NULL DEFAULT 0,
  paid_at            BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  UNIQUE KEY uk_app_account_token (app_account_token),
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
  invite_duration_days   INT          NOT NULL DEFAULT 0,
  ad_duration_days       INT          NOT NULL DEFAULT 0,
  default_allowed_count  INT          NOT NULL DEFAULT 0,
  logo                   VARCHAR(512) NOT NULL DEFAULT '',
  color                  VARCHAR(16)  NOT NULL DEFAULT '',
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
  PRIMARY KEY (redeemer_wx_id, code, feature_id)
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
		// 设备维邀请去重表：历史遗留；现网 care/growth 改用 InviteOncePerUser，不再写入本表。
		`CREATE TABLE IF NOT EXISTS feature_invite_device_grant (
  device_no    VARCHAR(64) NOT NULL,
  feature_id   VARCHAR(64) NOT NULL,
  code         VARCHAR(64) NOT NULL DEFAULT '',
  redeemer_wx_id BIGINT    NOT NULL DEFAULT 0,
  redeemed_at  BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (device_no, feature_id)
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
		// 账号维权益：activation_subject=user 的功能写入本表（care / growth）；与设备表并行，不改写旧唯一键。
		`CREATE TABLE IF NOT EXISTS feature_user_entitlement (
  id              BIGINT      NOT NULL AUTO_INCREMENT,
  wx_id           BIGINT      NOT NULL,
  feature_id      VARCHAR(64) NOT NULL,
  unlock_method   VARCHAR(32) NOT NULL DEFAULT '',
  expires_at      BIGINT      NOT NULL DEFAULT 0,
  quantity        INT         NOT NULL DEFAULT 0,
  source_ref      VARCHAR(128) NOT NULL DEFAULT '',
  created_at      BIGINT      NOT NULL DEFAULT 0,
  updated_at      BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wx_feature (wx_id, feature_id),
  KEY idx_wx (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		// 账号维免费试用记账：unused→used 仅在首次成功落库 claim 时翻转。
		`CREATE TABLE IF NOT EXISTS feature_user_trial (
  wx_id       BIGINT      NOT NULL,
  feature_id  VARCHAR(64) NOT NULL,
  status      VARCHAR(16) NOT NULL DEFAULT 'unused',
  used_at     BIGINT      NOT NULL DEFAULT 0,
  created_at  BIGINT      NOT NULL DEFAULT 0,
  updated_at  BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (wx_id, feature_id)
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
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  order_no           VARCHAR(64)  NOT NULL,
  device_no          VARCHAR(64)  NOT NULL,
  wx_id              BIGINT       NOT NULL DEFAULT 0,
  product_code       VARCHAR(64)  NOT NULL,
  channel            VARCHAR(32)  NOT NULL,
  amount_fen         INT          NOT NULL DEFAULT 0,
  currency           VARCHAR(8)   NOT NULL DEFAULT 'CNY',
  status             VARCHAR(16)  NOT NULL DEFAULT 'created',
  channel_txn_id     VARCHAR(128) NOT NULL DEFAULT '',
  app_account_token  CHAR(36)     NULL DEFAULT NULL COMMENT 'Apple StoreKit appAccountToken(UUID)；仅 apple_iap',
  grant_reason       VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Admin 手工授理由；支付单为空',
  created_at         BIGINT       NOT NULL DEFAULT 0,
  paid_at            BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  UNIQUE KEY uk_app_account_token (app_account_token),
  KEY idx_device_created (device_no, created_at),
  KEY idx_channel_txn (channel, channel_txn_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		// 喂养资格场景阈值（UCG / 值得留意）；Admin 可改，禁止随意新建 scene_key。
		`CREATE TABLE IF NOT EXISTS feeding_eligibility_scene (
  scene_key             VARCHAR(64) NOT NULL,
  required_days         INT         NOT NULL DEFAULT 1,
  min_records_per_day   INT         NOT NULL DEFAULT 10,
  updated_at            BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (scene_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		// 微信群二维码有效期（文件 er_code.png 在 gateway apk 目录）。
		`CREATE TABLE IF NOT EXISTS invite_group_qr (
  id          INT         NOT NULL,
  file_name   VARCHAR(64) NOT NULL DEFAULT 'er_code.png',
  expires_at  BIGINT      NOT NULL DEFAULT 0,
  updated_at  BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
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
		// Admin 手工授审查：授权理由；支付建单不写该列则保持默认空串。
		`ALTER TABLE vip_order ADD COLUMN grant_reason VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Admin 手工授理由；支付单为空'`,
		`ALTER TABLE feature_order ADD COLUMN grant_reason VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Admin 手工授理由；支付单为空'`,
		// Apple ASN：建单写入 UUID，通知用 appAccountToken 反查；可空以兼容历史/支付宝行。
		`ALTER TABLE vip_order ADD COLUMN app_account_token CHAR(36) NULL DEFAULT NULL COMMENT 'Apple appAccountToken UUID；仅 apple_iap'`,
		`ALTER TABLE feature_order ADD COLUMN app_account_token CHAR(36) NULL DEFAULT NULL COMMENT 'Apple appAccountToken UUID；仅 apple_iap'`,
	}
	for _, alterSQL := range alterCols {
		if _, err := db.Exec(ctx, alterSQL); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
				return err
			}
		}
	}
	// 唯一索引：MySQL 允许多个 NULL；重复执行忽略 Duplicate key name。
	for _, idxSQL := range []string{
		`ALTER TABLE vip_order ADD UNIQUE KEY uk_app_account_token (app_account_token)`,
		`ALTER TABLE feature_order ADD UNIQUE KEY uk_app_account_token (app_account_token)`,
	} {
		if _, err := db.Exec(ctx, idxSQL); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate") && !strings.Contains(msg, "1061") && !strings.Contains(msg, "already exists") {
				return err
			}
		}
	}
	// 邀请/广告授予天数列：仅首次加列时从 duration_days 回填，避免重启覆盖运维改数。
	inviteColAdded := false
	if _, err := db.Exec(ctx, `ALTER TABLE feature_def ADD COLUMN invite_duration_days INT NOT NULL DEFAULT 0`); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
			return err
		}
	} else {
		inviteColAdded = true
	}
	adColAdded := false
	if _, err := db.Exec(ctx, `ALTER TABLE feature_def ADD COLUMN ad_duration_days INT NOT NULL DEFAULT 0`); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
			return err
		}
	} else {
		adColAdded = true
	}
	if inviteColAdded {
		if _, err := db.Exec(ctx, `UPDATE feature_def SET invite_duration_days = duration_days`); err != nil {
			return err
		}
	}
	if adColAdded {
		if _, err := db.Exec(ctx, `UPDATE feature_def SET ad_duration_days = duration_days`); err != nil {
			return err
		}
	}
	// 开通主体：device=对机全家共享；user=对人一人一份。首次加列回填 device，并将成长轨迹定为 user。
	subjectColAdded := false
	if _, err := db.Exec(ctx, `ALTER TABLE feature_def ADD COLUMN activation_subject VARCHAR(16) NOT NULL DEFAULT 'device'`); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
			return err
		}
	} else {
		subjectColAdded = true
	}
	if subjectColAdded {
		if _, err := db.Exec(ctx, `UPDATE feature_def SET activation_subject='user' WHERE feature_id=?`, FeatureIDGrowthTrajectoryPredict); err != nil {
			return err
		}
	}
	// 每次 Ensure：care / growth 均为账号维开通。
	if _, err := db.Exec(ctx, `UPDATE feature_def SET activation_subject='user' WHERE feature_id IN (?,?)`,
		FeatureIDCareAlertSmartRemind, FeatureIDGrowthTrajectoryPredict); err != nil {
		return err
	}
	// 剥离 ad 开通方式（care / growth）。预测槽位不再在启动时写回。
	if _, err := db.Exec(ctx, `UPDATE feature_def SET unlock_methods='payment,invite_code', invite_duration_days=IF(invite_duration_days<=0,7,invite_duration_days), updated_at=? WHERE feature_id IN (?,?)`,
		time.Now().Unix(), FeatureIDCareAlertSmartRemind, FeatureIDGrowthTrajectoryPredict); err != nil {
		return err
	}
	// 功能视觉：logo（OSS objectKey）与主色 hex；已有库补列。
	for _, alterSQL := range []string{
		`ALTER TABLE feature_def ADD COLUMN logo VARCHAR(512) NOT NULL DEFAULT ''`,
		`ALTER TABLE feature_def ADD COLUMN color VARCHAR(16) NOT NULL DEFAULT ''`,
	} {
		if _, err := db.Exec(ctx, alterSQL); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
				return err
			}
		}
	}
	// 邀请去重键升级为人×码×功能（未发布环境可接受失败重试）。
	if _, err := db.Exec(ctx, `ALTER TABLE feature_invite_feature_grant DROP PRIMARY KEY, ADD PRIMARY KEY (redeemer_wx_id, code, feature_id)`); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "Multiple primary key") && !strings.Contains(msg, "1068") &&
			!strings.Contains(msg, "Duplicate") && !strings.Contains(msg, "already exists") {
			// 已是新主键或表空时部分环境报错信息不同：仅记录，不阻断若已含 code 的复合主键。
			g.Log().Warningf(ctx, "[cash-schema] invite grant PK migrate: %v", err)
		}
	}
	now := time.Now().Unix()
	// 种子 Apple 商品 ID 为空；Admin/SQL 已写入的非空值不覆盖。价格同样不覆盖已有 price_fen。
	_, err := db.Exec(ctx, `
INSERT INTO vip_product (product_code, title, price_fen, original_price_fen, duration_days, apple_product_id, status, updated_at)
VALUES (?, 'VIP月会员', ?, ?, ?, '', 1, ?)
ON DUPLICATE KEY UPDATE
  title=VALUES(title),
  duration_days=VALUES(duration_days),
  apple_product_id=IF(VALUES(apple_product_id)='', apple_product_id, VALUES(apple_product_id)),
  original_price_fen=IF(original_price_fen=0, VALUES(original_price_fen), original_price_fen),
  status=1,
  updated_at=VALUES(updated_at)`,
		ProductMonthly19, ProductPriceFen, ProductOriginalPriceFen, ProductDurationD, now)
	if err != nil {
		return err
	}
	// 值得留意：账号维；支付/邀请；邀请 7 天；默认主色青绿。
	_, err = db.Exec(ctx, `
INSERT INTO feature_def (feature_id, title, description, unlock_methods, duration_days, invite_duration_days, ad_duration_days, default_allowed_count, activation_subject, logo, color, status, sort_order, updated_at)
VALUES (?, '值得留意智能提醒', '开通后可查看值得留意智能提醒', 'payment,invite_code', 7, 7, 0, 0, 'user', '', '#0D9488', 1, 20, ?)
ON DUPLICATE KEY UPDATE
  activation_subject='user',
  unlock_methods='payment,invite_code',
  invite_duration_days=IF(invite_duration_days<=0, VALUES(invite_duration_days), invite_duration_days),
  color=IF(color='' OR color IS NULL, VALUES(color), color),
  updated_at=VALUES(updated_at)`,
		FeatureIDCareAlertSmartRemind, now)
	if err != nil {
		return err
	}
	// 值得留意付费 30 天 SKU；INSERT IGNORE 不覆盖运维改价。
	_, err = db.Exec(ctx, `
INSERT IGNORE INTO feature_product (product_code, feature_id, grant_kind, grant_quantity, price_fen, original_price_fen, duration_days, apple_product_id, status, updated_at)
VALUES (?, ?, 'entitlement', 1, 990, 0, 30, '', 1, ?)`,
		CareAlertSmartRemindProductCode, FeatureIDCareAlertSmartRemind, now)
	if err != nil {
		return err
	}
	// 停用旧永久 SKU；若运维已把旧码改成 30d 则仅保证 duration。
	_, _ = db.Exec(ctx, `UPDATE feature_product SET status=0, updated_at=? WHERE product_code=?`,
		now, CareAlertSmartRemindProductCodeLegacyPerm)
	_, _ = db.Exec(ctx, `UPDATE feature_product SET duration_days=30, updated_at=? WHERE product_code=? AND duration_days=0 AND status=1`,
		now, CareAlertSmartRemindProductCode)
	// 成长轨迹：账号维；支付/邀请；邀请 7 天；默认主色橙。
	_, err = db.Exec(ctx, `
INSERT INTO feature_def (feature_id, title, description, unlock_methods, duration_days, invite_duration_days, ad_duration_days, default_allowed_count, activation_subject, logo, color, status, sort_order, updated_at)
VALUES (?, '成长轨迹预测', '开通后可使用成长轨迹预测', 'payment,invite_code', 7, 7, 0, 0, 'user', '', '#EA580C', 1, 30, ?)
ON DUPLICATE KEY UPDATE
  activation_subject='user',
  unlock_methods='payment,invite_code',
  invite_duration_days=IF(invite_duration_days<=0, VALUES(invite_duration_days), invite_duration_days),
  color=IF(color='' OR color IS NULL, VALUES(color), color),
  updated_at=VALUES(updated_at)`,
		FeatureIDGrowthTrajectoryPredict, now)
	if err != nil {
		return err
	}
	// 成长轨迹付费 30 天 SKU（1900 分）；INSERT IGNORE 不覆盖运维改价；无永久 SKU。
	_, err = db.Exec(ctx, `
INSERT IGNORE INTO feature_product (product_code, feature_id, grant_kind, grant_quantity, price_fen, original_price_fen, duration_days, apple_product_id, status, updated_at)
VALUES (?, ?, 'entitlement', 1, 1900, 0, 30, '', 1, ?)`,
		GrowthTrajectoryPredictProductCode, FeatureIDGrowthTrajectoryPredict, now)
	if err != nil {
		return err
	}
	// 喂养资格场景种子：仅 INSERT 忽略已存在，不覆盖运维已改阈值。
	_, err = db.Exec(ctx, `
INSERT IGNORE INTO feeding_eligibility_scene (scene_key, required_days, min_records_per_day, updated_at)
VALUES (?, 7, 10, ?), (?, 2, 10, ?)`,
		SceneKeyUCGEntry, now, SceneKeyCareAlertEntry, now)
	return err
}

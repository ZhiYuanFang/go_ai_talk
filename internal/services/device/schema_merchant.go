package device

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	WxRoleParent = "parent"
	WxRoleStaff  = "staff"

	BabyMemberSourceSelf           = "self"
	BabyMemberSourceBind           = "bind"
	BabyMemberSourceMerchantAssign = "merchant_assign"
)

// EnsureMerchantConsoleSchema 幂等创建商户/成员表并为 wx 补 role 列（device 库）。
func EnsureMerchantConsoleSchema(ctx context.Context) error {
	if err := ensureWxRoleColumn(ctx); err != nil {
		return err
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS merchant (
  id            BIGINT       NOT NULL AUTO_INCREMENT,
  store_name    VARCHAR(64)  NOT NULL COMMENT '店名，全局唯一',
  password_hash VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'bcrypt',
  created_at    BIGINT       NOT NULL DEFAULT 0,
  updated_at    BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_name (store_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='B 端商户主账号'`,
		`CREATE TABLE IF NOT EXISTS wx_baby_member (
  id         BIGINT      NOT NULL AUTO_INCREMENT,
  wx_id      BIGINT      NOT NULL,
  device_no  VARCHAR(64) NOT NULL,
  source     VARCHAR(32) NOT NULL DEFAULT 'self' COMMENT 'self|merchant_assign|bind',
  created_at BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wx_device (wx_id, device_no),
  KEY idx_device (device_no),
  KEY idx_wx (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账号↔宝宝成员真源'`,
		`CREATE TABLE IF NOT EXISTS merchant_baby (
  id           BIGINT      NOT NULL AUTO_INCREMENT,
  merchant_id  BIGINT      NOT NULL,
  device_no    VARCHAR(64) NOT NULL,
  created_at   BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_device (device_no),
  KEY idx_merchant (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户拥有的虚拟宝宝'`,
	}
	for _, sql := range stmts {
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			return err
		}
	}
	glog.Infof(ctx, "[device-schema] merchant console tables ensured")
	return nil
}

func ensureWxRoleColumn(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'wx' AND COLUMN_NAME = 'role'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
ALTER TABLE wx ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'parent' COMMENT 'parent|staff 账号级单值' AFTER password`)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "Duplicate column") || strings.Contains(msg, "1060") {
			return nil
		}
		return err
	}
	glog.Infof(ctx, "[device-schema] wx.role 列已添加")
	return nil
}

// EnsureBackfillWxBabyMembers 将非空 wx.device_no 回填为成员表一行（幂等 INSERT IGNORE）。
func EnsureBackfillWxBabyMembers(ctx context.Context) error {
	now := time.Now().Unix()
	res, err := g.DB().Exec(ctx, `
INSERT IGNORE INTO wx_baby_member (wx_id, device_no, source, created_at)
SELECT id, device_no, ?, ?
FROM wx
WHERE device_no IS NOT NULL AND TRIM(device_no) <> ''`, BabyMemberSourceSelf, now)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		glog.Infof(ctx, "[device-schema] wx_baby_member 回填 %d 行", n)
	}
	return nil
}

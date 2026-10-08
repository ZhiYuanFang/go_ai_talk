package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureXiaozhiMcpBindingTable 启动时保证 xiaozhi_mcp_binding 表存在（幂等）。
// 业务：一人多小智音箱；mcp_token 全局唯一；speaker_mac 全局唯一；status=1 为 active。
func EnsureXiaozhiMcpBindingTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'xiaozhi_mcp_binding'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE xiaozhi_mcp_binding (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
  wx_id BIGINT NOT NULL DEFAULT 0 COMMENT '归属 wx 主键',
  device_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT '喂养落点宝宝设备号',
  mcp_token VARCHAR(512) NOT NULL DEFAULT '' COMMENT '小智 MCP 接入点 token',
  speaker_mac VARCHAR(32) NOT NULL DEFAULT '' COMMENT '音箱 MAC，规范化小写冒号分隔',
  alias VARCHAR(128) NOT NULL DEFAULT '' COMMENT '用户备注名',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1=active 0=disabled',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建 unix 秒',
  updated_at BIGINT NOT NULL DEFAULT 0 COMMENT '更新 unix 秒',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_token (mcp_token),
  UNIQUE KEY uk_speaker_mac (speaker_mac),
  KEY idx_wx_id (wx_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小智 MCP 音箱绑定'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] xiaozhi_mcp_binding 表已创建")
	return nil
}

// EnsureXiaozhiMcpSpeakerMacColumn 为已有表补齐 speaker_mac 列与全局唯一索引（幂等）。
// 业务：空 MAC 历史测试行在加唯一索引前删除，避免 uk 冲突；新写入禁止空 MAC。
//
// Side Effects: ALTER / DELETE / CREATE INDEX。
func EnsureXiaozhiMcpSpeakerMacColumn(ctx context.Context) error {
	col, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'xiaozhi_mcp_binding' AND COLUMN_NAME = 'speaker_mac'`)
	if err != nil {
		return err
	}
	if col.Int() == 0 {
		_, err = g.DB().Exec(ctx, `
ALTER TABLE xiaozhi_mcp_binding
  ADD COLUMN speaker_mac VARCHAR(32) NOT NULL DEFAULT '' COMMENT '音箱 MAC，规范化小写冒号分隔' AFTER mcp_token`)
		if err != nil {
			return err
		}
		glog.Infof(ctx, "[device-schema] xiaozhi_mcp_binding.speaker_mac 列已添加")
	}
	// 加唯一索引前清理空 MAC，避免多行空串撞 uk。
	affected, err := g.DB().Exec(ctx, `DELETE FROM xiaozhi_mcp_binding WHERE speaker_mac = '' OR speaker_mac IS NULL`)
	if err != nil {
		return err
	}
	if n, _ := affected.RowsAffected(); n > 0 {
		glog.Warningf(ctx, "[device-schema] 已删除无 speaker_mac 的小智绑定行 count=%d", n)
	}
	idx, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'xiaozhi_mcp_binding' AND INDEX_NAME = 'uk_speaker_mac'`)
	if err != nil {
		return err
	}
	if idx.Int() == 0 {
		_, err = g.DB().Exec(ctx, `ALTER TABLE xiaozhi_mcp_binding ADD UNIQUE KEY uk_speaker_mac (speaker_mac)`)
		if err != nil {
			return err
		}
		glog.Infof(ctx, "[device-schema] xiaozhi_mcp_binding.uk_speaker_mac 已创建")
	}
	return nil
}

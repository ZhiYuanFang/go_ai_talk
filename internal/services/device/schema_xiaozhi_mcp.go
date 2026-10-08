package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureXiaozhiMcpBindingTable 启动时保证 xiaozhi_mcp_binding 表存在（幂等）。
// 业务：一人多小智音箱；mcp_token 全局唯一；status=1 为 active。
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
  alias VARCHAR(128) NOT NULL DEFAULT '' COMMENT '用户备注名',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1=active 0=disabled',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建 unix 秒',
  updated_at BIGINT NOT NULL DEFAULT 0 COMMENT '更新 unix 秒',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_token (mcp_token),
  KEY idx_wx_id (wx_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小智 MCP 音箱绑定'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] xiaozhi_mcp_binding 表已创建")
	return nil
}

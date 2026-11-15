package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsurePredictProIdeaTables 启动时幂等创建预测 Pro 想法表与点赞表。
// 业务说明：想法墙独立于 feedback；库为 device 域 ai_voice_device。
func EnsurePredictProIdeaTables(ctx context.Context) error {
	if err := ensurePredictProIdeaTable(ctx); err != nil {
		return err
	}
	return ensurePredictProIdeaLikeTable(ctx)
}

func ensurePredictProIdeaTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'predict_pro_idea'`)
	if err != nil {
		return err
	}
	// 业务逻辑：表已存在则跳过，不做 ALTER；查询侧兼容 deleted_at=0 与 NULL。
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE predict_pro_idea (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '想法主键',
  wx_id BIGINT NOT NULL COMMENT '提交者 wx 主键',
  device_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提交时绑定设备号',
  baby_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提交时宝宝名快照',
  content VARCHAR(200) NOT NULL COMMENT '用户原文',
  display_text VARCHAR(200) NOT NULL COMMENT '对外展示文案（运营可改）',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建 unix 秒',
  updated_at BIGINT NOT NULL DEFAULT 0 COMMENT '更新 unix 秒',
  deleted_at BIGINT NOT NULL DEFAULT 0 COMMENT '软删 unix 秒；0=未删；>0=已删',
  PRIMARY KEY (id),
  KEY idx_ppi_wx_created (wx_id, created_at),
  KEY idx_ppi_created (created_at),
  KEY idx_ppi_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预测 Pro 想法'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] predict_pro_idea 表已创建")
	return nil
}

func ensurePredictProIdeaLikeTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'predict_pro_idea_like'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE predict_pro_idea_like (
  idea_id BIGINT NOT NULL COMMENT '想法 id',
  wx_id BIGINT NOT NULL COMMENT '点赞者 wx 主键',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT '点赞 unix 秒',
  PRIMARY KEY (idea_id, wx_id),
  KEY idx_ppil_wx (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预测 Pro 想法点赞'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] predict_pro_idea_like 表已创建")
	return nil
}

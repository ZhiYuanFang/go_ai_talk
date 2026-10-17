package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureSurveyTables 启动时保证满意度问卷三张表存在（幂等）。
// 业务：活动开关与上海时区起止（unix 秒）；选项行；同一活动同一 wx 只留一条提交，选项文案以快照保存。
func EnsureSurveyTables(ctx context.Context) error {
	if err := ensureSurveyCampaignTable(ctx); err != nil {
		return err
	}
	if err := ensureSurveyOptionTable(ctx); err != nil {
		return err
	}
	return ensureSurveyResponseTable(ctx)
}

func ensureSurveyCampaignTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_campaign'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE survey_campaign (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '活动主键；App 只读 id 最大的一轮',
  enabled TINYINT NOT NULL DEFAULT 0 COMMENT '1=开启；0 时即使处于有效期也不展示',
  start_at BIGINT NOT NULL DEFAULT 0 COMMENT '有效期开始 unix 秒（Asia/Shanghai 墙钟换算）',
  end_at BIGINT NOT NULL DEFAULT 0 COMMENT '有效期结束 unix 秒（含该秒）',
  intro VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'sheet 说明文案',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建 unix 秒',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='满意度问卷活动'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] survey_campaign 表已创建")
	return nil
}

func ensureSurveyOptionTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_option'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE survey_option (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '选项主键',
  campaign_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属活动',
  text VARCHAR(200) NOT NULL DEFAULT '' COMMENT '满意点文案',
  sort INT NOT NULL DEFAULT 0 COMMENT '展示顺序，小在前',
  enabled TINYINT NOT NULL DEFAULT 1 COMMENT '1=对 App 展示',
  PRIMARY KEY (id),
  KEY idx_campaign_sort (campaign_id, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='满意度问卷选项行'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] survey_option 表已创建")
	return nil
}

func ensureSurveyResponseTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE survey_response (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '提交主键',
  campaign_id BIGINT NOT NULL DEFAULT 0 COMMENT '活动 id',
  wx_id BIGINT NOT NULL DEFAULT 0 COMMENT '提交用户',
  device_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提交时会话设备号',
  baby_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提交时宝宝昵称快照',
  option_snapshot TEXT COMMENT '所选选项提交时文案 JSON 数组',
  comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '更多意见',
  submitted_at BIGINT NOT NULL DEFAULT 0 COMMENT '提交 unix 秒',
  PRIMARY KEY (id),
  UNIQUE KEY uk_campaign_wx (campaign_id, wx_id),
  KEY idx_campaign_time (campaign_id, submitted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='满意度问卷提交（一次活动一人一行）'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] survey_response 表已创建")
	return nil
}

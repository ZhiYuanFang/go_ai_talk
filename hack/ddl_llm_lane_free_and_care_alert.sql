-- VIP∪额度共同策略：llm_lane_config 增加 free 列与 careAlert 种子行依赖列。
-- voice-service EnsureLLMLaneSchema 启动时亦会幂等执行等价 ALTER。

ALTER TABLE llm_lane_config
  ADD COLUMN free_provider VARCHAR(32) NOT NULL DEFAULT '' COMMENT '额度不足模型供应商；空=omit' AFTER model,
  ADD COLUMN free_model VARCHAR(64) NOT NULL DEFAULT '' COMMENT '额度不足模型名；空=omit' AFTER free_provider;

ALTER TABLE ai_quota_default
  ADD COLUMN care_alert_monthly_limit INT NOT NULL DEFAULT 10 AFTER clinic_ai_monthly_limit;

ALTER TABLE ai_quota_user_override
  ADD COLUMN care_alert_monthly_limit INT NULL AFTER clinic_ai_monthly_limit;

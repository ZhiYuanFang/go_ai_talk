# feedback 用户反馈表迁移说明

## 目标

- 在 device 域默认库（`ai_voice_device` / 测试环境 `ai_voice_device_test`）创建 `feedback` 表
- 支撑 App 用户提交反馈、查看历史与 Admin 单次官方回复

## 前置检查

1. 确认数据库连接为 `DEVICE_DB_LINK` / `database.default` 指向的 device 库
2. 确认 `wx` 表已存在（`feedback.wx_id` 关联 `wx.id`）

## DDL 执行

在目标库执行 [hack/ddl_feedback.sql](../../hack/ddl_feedback.sql)。

```bash
# 示例（按实际连接串调整）
mysql -h <host> -u <user> -p ai_voice_device < hack/ddl_feedback.sql
mysql -h <host> -u <user> -p ai_voice_device_test < hack/ddl_feedback.sql
```

## 发布顺序

1. 低峰/维护窗口执行 DDL
2. 部署 device-service（含 App/Admin feedback API）
3. 验证 `/device/admin/feedback-records` 与 API
4. 发布 Flutter 客户端

## 回滚

- 可下线 Admin 入口与 App 设置入口；不建议删除表以免丢失用户反馈数据

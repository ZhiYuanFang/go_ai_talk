## 1. voice 代理 API

- [x] 1.1 `python_ai_client.go`：实现 `ListRuleGaps(ctx, offset, limit, dimension, status)` → `GET /v1/admin/rule-gaps`；`PatchRuleGapStatus(ctx, id, status)` → `PATCH /v1/admin/rule-gaps/{id}`
- [x] 1.2 `api/v1`：新增 `GET /voice/admin/api/rule-gaps`、`PATCH /voice/admin/api/rule-gaps/{id}` 请求/响应 DTO（字段对齐 Python：id/dimension/segmentText|segment_text 等，实现时统一 camel 映射）
- [x] 1.3 新增 `VoiceAdminRuleGapsCtrl`（口令同 intent-vectors）；在 `register_voice_service` Bind

## 2. Hub 静态页与导航

- [x] 2.1 新增 `resource/public/rule-gaps-admin.html`：筛选（dimension/status）、分页列表（整句/段/reason/llm/device/时间/状态）、行内「标已处理」
- [x] 2.2 登记 `admin-modules.js`（标题「规则缺口」，与意图向量并列）、`admin_static_pages.go`、`gateway_app_auth_exempt.go`（及 usage skip 静态路径若 intent-vector 有同类登记则对齐）
- [x] 2.3 页面 hint：数据在 Python `rule_gaps` 收件箱；改规则须 Python 发版，非热更新

## 3. 自检

- [x] 3.1 确认路径落在 `/voice/admin/api/*` 反代内；无 App 新路径、不改 `maintenance_skip` 业务策略
- [x] 3.2 跑 `hack/check-service-import`；相关包可编译
- [x] 3.3 联调：Python 有 gaps 时 Hub 可见；标 dismissed 后默认 open 列表消失；Python 宕机时有明确错误

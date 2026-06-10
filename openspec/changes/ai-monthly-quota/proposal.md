## Why

AI 润笔与语音喂养 AI 对话目前无月度使用上限，存在成本不可控与滥用风险。产品需要按用户（wxId）对两个功能分别限次：默认各 5 次/自然月，运营可在管理后台独立调整全局默认（如润笔 10、喂养 AI 5）及单人 override；额度用尽时 App 弹框「本月额度已用完」。wxId=0 不允许使用 AI；扣减仅在 AI 成功返回后发生。

## What Changes

- 在 **device-service** 新增 AI 月度额度权威：全局默认配置（`polish`、`voice_ai` 两个独立字段，初始各 5）、per-wxId override、Redis 按月用量计数（`Asia/Shanghai` 自然月 `YYYYMM`）。
- 新增 **internal** 契约：`POST /device/internal/api/ai-quota/check`（预检，不扣减）、`POST /device/internal/api/ai-quota/consume`（成功扣减）；ucg-service 与 voice-service MUST 经 HTTP 调用，禁止跨库直查。
- 新增 **App** API：`GET /device/app/api/ai-quota`（当前用户两功能 used/limit，供 Flutter 可选展示）。
- 新增 **Admin** API：`GET/PUT /device/admin/api/ai-quota/default`（全局默认）、`GET/PUT /device/admin/api/ai-quota/user`（按 wxId override）；鉴权复用 `X-Admin-Password`。
- **ucg-service**：`POST /ucg/app/api/posts/polish` 在调 DashScope 前预检额度，成功返回正文后扣减；超额返回明确业务错误供 Flutter 弹框。
- **voice-service**：在触发 LLM（母婴 DeepSeek / casual 流式）前预检 wxId>0 与额度；LLM 成功完成后扣减；模式切换、规则回复、LLM 失败兜底不扣减。
- 扩展 **`ucg-admin.html`**「AI 配置」Tab：全局润笔/喂养 AI 默认次数 + wxId 单人 override 表单。
- **Flutter**（`flutter_ai_talk` 仓库）：识别额度耗尽错误并弹框「本月额度已用完」；wxId=0 引导登录（非额度文案）。

## Capabilities

### New Capabilities

- `ai-monthly-quota`：AI 功能月度额度（device 权威、Redis 用量、internal/app/admin API、ucg 润笔与 voice 喂养 AI 接入、管理页、Flutter 错误契约）。

### Modified Capabilities

（无。不修改 gateway-app-api-usage-stats 等既有统计行为；额度 enforcement 与被动观测分离。）

## Impact

- **服务**：`device-service`（配置表、Redis、internal/admin/app API）；`ucg-service`（润笔前/后额度）；`voice-service`（WS LLM 前/后额度，经 device HTTP）；`gateway-app-server`（反代 device admin/app 路径，无新采集逻辑）。
- **静态页**：`resource/public/ucg-admin.html`（或 device admin 链入同一配置区）。
- **数据库**：device 库新增 `ai_quota_default`（singleton）、`ai_quota_user_override`（wx_id 唯一）；无 ucg/voice 库迁移。
- **Redis**：device 进程 `ai:usage:{feature}:{wxId}:{YYYYMM}`，TTL 约 90 天。
- **边界**：`/voice/asr/ws` 纯听写不计入；功能使用统计模块不参与限流。

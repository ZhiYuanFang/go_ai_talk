## Why

润笔与胖宝诊疗在月度额度用尽后当前直接返回 **40302** 硬阻断，用户无法继续使用 AI 能力。智谱免费 flash 模型（`glm-4.6v-flash` / `glm-4.1v-thinking-flash`）已接入 aimodel lane 且不限调用次数，应在额度用尽时自动降级至免费路径，提升可用性；喂养对话（`voice_ai`）仍保持额度硬限制。客户端需明确展示「已降速」状态，避免用户误以为功能完全不可用。

## What Changes

- **润笔（polish）**：`used >= limit` 时 **不再** 返回 HTTP 40302；改走 **degraded** 路径，强制经 `LanePolish` 调用智谱 **`glm-4.6v-flash`**，**不** consume 月度额度；成功响应 MAY 含 `quotaDegraded: true`。
- **胖宝诊疗（clinic_ai）**：仅因额度用尽时 **不再** 返回 WS 40302；改走 **degraded** 路径，经 `LaneClinic` 调用智谱 **`glm-4.1v-thinking-flash`**，**不** consume `clinic_ai`；40301（未登录）、42901（限流）、50301（队列满）等行为不变。
- **喂养对话（voice_ai）**：**不变**——额度用尽仍返回 WS 40302，无 degraded fallback。
- **额度读 API**：`GET /ucg/app/api/ai-quota` 与 `GET /voice/app/api/ai-quota` 各 feature 快照新增 **`degraded: bool`**（当 `used >= limit` 时为 `true`）。
- **Flutter（flutter_ai_talk）**：`remaining=0` 且 `degraded=true` 时展示醒目降速文案（润笔：「本月润笔额度已用完，已降速」；胖宝：「本月胖宝诊疗额度已用完，已降速」）；`voice_ai` 仍弹 40302「本月额度已用完」；润笔响应 `quotaDegraded` 可选 toast。
- **BREAKING（客户端语义）**：润笔 HTTP 与胖宝 WS 在**仅额度用尽**场景不再返回 40302；依赖该码判断「功能完全不可用」的旧客户端需升级。

## Capabilities

### New Capabilities

- `app-ai-quota-degraded-ui`：Flutter 额度展示与 40302 弹框语义——degraded 降速文案、`degraded` 字段解析、润笔 `quotaDegraded` 可选提示；`voice_ai` 40302 行为不变。

### Modified Capabilities

- `ai-monthly-quota`：**BREAKING**——润笔/clinic 额度用尽改 degraded 路径、不 40302；voice_ai 40302 不变；Flutter 40302 弹框范围收窄至 voice_ai（及非额度类 40302）。
- `ucg-ai-quota`：`degraded` 字段；润笔双路径（正常 consume / degraded 不 consume）；`quotaDegraded` 响应字段。
- `pangbao-ai-clinic`：clinic degraded 路径（额度用尽仍可调 LLM、不 consume clinic_ai）。

## Impact

- **后端（go_ai_talk）**：`internal/controller/ucg_app_api.go`（`PostsPolish`、`AIQuotaGet`）；`internal/services/ucg/ai_quota.go`、`compose_ai.go`；`internal/services/voice/clinic_service.go`、`ai_quota.go`；`internal/services/contracts/ai_quota_contracts.go`；`api/v1/ucg_app_http.go`、`api/v1/voice_app_http.go`（或等价 quota 响应 DTO）。
- **前端（flutter_ai_talk）**：`app/lib/data/ai_quota_models.dart`；`app/lib/ui/widgets/ai_quota_remaining_hint.dart`；`app/lib/ui/ucg/ucg_compose_screen.dart`；`app/lib/ui/pangbao/pangbao_ai_screen.dart`；润笔 API 响应解析。
- **不变**：Admin 配额配置、Redis 用量键、`voice_ai` guard/consume 逻辑、40301/50301/42901 语义。
- **App API 使用统计**：无新增 App HTTP 路由；`degraded` 为既有 quota 读 API 字段扩展，**不计入** usage 统计变更（与现有 `GET */ai-quota` 维护型排除一致）。

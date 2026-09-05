## Why

Flutter 已停止调用 `/device/tip/generate` 与 tip/clinic 点赞飞轮；运维页仍残留 tip SSE 调试入口。服务端继续保留 TipStream、Clinic/Tip feedback 与对应 gateway 反代，增加 voice 进程杂货面与维护成本。喂养/诊疗连续对话已自行消化反馈意图，HTTP 飞轮不再需要。值得留意（care-alert）仍为活跃产品，**其飞轮与宿主均保留在 voice**。

## What Changes

- **BREAKING**：下线 `POST /device/tip/generate`（SSE 小贴士）；卸载 `TipCtrl`、`VoiceService.TipStream`、Python tip 客户端及相关 contracts。
- **BREAKING**：下线 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`；删除 `DeviceClinicFeedbackController`（及 tip feedback 契约）。
- gateway-app：移除对 `/device/tip/*`、`/device/api/tip/*`、`/device/api/clinic/*`（仅反馈用途）的 voice 反代绑定；**保留** `/device/api/care-alert/*`。
- 运维：删除 `resource/public/history.html`（或等价运维页）中对 tip generate 的调用与 UI。
- usage / Bearer：清理 tip/clinic feedback 相关 `maintenance_skip` 条目与过时注释；tip generate 随路由消失后无需再讨论统计策略。
- **非目标 / 明确保留**：
  - care-alert：`GET|DELETE /device/api/care-alert/*` 与 **`POST /device/api/care-alert/feedback` 飞轮** 全部保留；宿主仍为 **voice-service**；配额/lane 不迁 device。
  - clinic WS：`/voice/clinic/ws` 及 `clinic_ai` 额度保留。
  - 不改 Flutter（客户端已无 tip/clinic feedback 调用）；不新建微服务；不新增测试文件。

## Capabilities

### New Capabilities

- `remove-device-tip`：服务端与运维彻底移除小贴士 SSE 生成能力及 tip Python/contracts 接线。
- `remove-clinic-tip-feedback`：移除 clinic/tip HTTP 点赞飞轮；gateway 反代与 usage skip 同步清理；**不得**删除 care-alert feedback。

### Modified Capabilities

- （相对 `openspec/specs/v3.0.0/spec.md` 中 tip generate、clinic/tip feedback、usage skip 等 Requirement：本变更 specs 以「MUST NOT 再暴露」取代基线「SHALL 暴露」；归档合并时再写入新版基线。本 change 目录内不直接改 v3.0.0 文件。）

## Impact

- **进程**：`voice-service`（删 tip/feedback 注册与实现）；`gateway-app-server`（反代模式、usage skip）。
- **API**：**BREAKING** 对仍调用 tip/clinic feedback 的旧客户端或脚本；App Flutter 主路径已停用；运维 tip 调试同步删除。
- **保留面**：care-alert 全路径 + care-alert feedback；clinic WS；voice/clinic AI 额度。
- **库 / Redis**：无 schema 变更；不改 care-alert 日缓存键。
- **文档**：若 runbook 仍写 tip 运维调试，实现时按需删改相关段落（tasks 勾选）。

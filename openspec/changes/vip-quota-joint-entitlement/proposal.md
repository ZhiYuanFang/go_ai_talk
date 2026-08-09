## Why

账号 VIP 与月度 AI 额度目前割裂：额度路径用尽后靠代码硬编码 degraded 模型，care-alert 仅按 VIP 硬切 DeepSeek/Zhipu 且无独立额度，硬件直连语音与 MCP 也未纳入统一权益。需要把 **VIP、额度、硬件特权** 收成同一套 premium 决策，并把「正式模 / 免费模」交给 Admin 可配，避免业务各自为政。

## What Changes

- 引入统一权益：`isPremium = isVip(wxId) ∨ hardwarePrivilege ∨ quota.Allowed(feature)`；对外额度快照在 VIP 时视为有额度；**VIP 与硬件特权路径不计次**（consume 跳过）。
- 选模：premium 使用各业务 lane 正式模型；非 premium 使用该 lane 可配的「额度不足/免费」模型；**免费模型可为空**，此时 Go 不向 Python 传 model，由 Python 自选。
- **care-alert**：新增独立额度 feature `care_alert` 与独立 LLM lane `careAlert`（含并发闸门）；选模改走统一权益（非 VIP 但有额度也走正式模）；生成成功时 **仅非 VIP 扣 `care_alert` 额度**；废除 VIP→DeepSeek / 非 VIP→Zhipu 硬切。
- **polish / tip** 与喂养/诊疗同一套 VIP∪额度选模策略；tip **默认挂靠** `clinic_ai` feature 与 `clinic` lane（含 free 段）。
- Admin「AI 模型与并发」：Voice 增加 `careAlert`；对 voiceUnderstanding / clinic / careAlert / polish 增加 free 配置；**Sim 四条不配 free**。
- 硬件特权保留：`/voice/chat/ws`（gateway 设备直连）与 MCP→internal text chat 入口打标后按 premium、不计次。
- 逐步停用代码内 `Degraded*Profile` 作为选模真相源（改读 lane.free）。

## Capabilities

### New Capabilities

- `vip-quota-entitlement`：VIP∪额度∪硬件特权的公共判定、VIP/硬件不计次、原子选模出口与 Python omit 语义。
- `care-alert-quota-lane`：`care_alert` 独立额度、`careAlert` lane/并发、生成成功非 VIP 计次，以及相对旧 VIP-only 选模的替换。
- `lane-free-model`：各业务 lane 的免费模型配置（可空→Python）、Admin API/UI 字段约定（不含 Sim）。

### Modified Capabilities

（复用/变更 v3.0.0 基线中的 capability 边界）

- `voice-ai-quota`：额度快照并入 VIP；硬件特权；tip 挂 clinic；degraded 选模改为 free 配置。
- `ucg-ai-quota`：polish 并入 VIP∪额度与 free 选模；VIP 不计次。
- `llm-lane-admin` / `llm-lane-gate`：新增 `careAlert` lane 与闸门；正式模字段旁增加 free；Sim 无 free。
- `ai-model-admin-ui`：页面展示 careAlert 与 free 编辑；仍为运维页、不计入 App usage。
- `app-ai-quota-degraded-ui`：App 额度展示与 VIP「有额度/不计次」语义对齐（跨仓 Flutter 可跟；本仓至少保证 API 字段语义一致）。

## Impact

- **进程**：`voice-service`（额度/选模/care-alert/tip/clinic/voice）、`ucg-service`（polish）、`mcp-service` / gateway 语音 WS 入口打标、`gateway-app` 静态 Admin 页；cash-service 仅继续提供 VIP 读契约。
- **库表**：voice 域 `llm_lane_config` 扩展 free 列（及 `careAlert` 种子行）；`care_alert` 额度默认/override 存储与现有 voice AI 额度同族模式。
- **API**：`/voice/admin/api/llm-lanes`、`/voice/admin/api/ai-quota*`、`/ucg/admin/api/*` / App `ai-quota`、care-alert 生成路径；Python 请求体 model 可选。
- **跨仓**：Flutter 额度/VIP 文案宜跟进；本变更以 Go API 语义为准，可单独立项改 UI。
- **基线**：supersede care-alert「仅 VIP 选 DeepSeek/Zhipu」及硬编码 degraded 选模相关 Requirement（见 v3.0.0 `voice-ai-quota` / care-alert 相关章节）。

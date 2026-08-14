## Why

自然语言喂养当前存在双外壳（`/voice/chat/ws` 与 history `chat`/`chat/stream`→voice internal text），特权与额度语义分叉，运维与产品都难对齐。Flutter 产品侧已切离外壳 B；需要在 Go 侧收敛为唯一 WS 入口，并去掉喂养「额尽 free 选模」与额度页喂养项。同时，设备数据运维页缺少仿 App 的 tip / 值得留意 / 横屏文字对话调试能力，且 care-alert 日缓存导致无法强刷重生。

## What Changes

- **唯一喂养入口**：自然语言喂养仅保留 `/voice/chat/ws`（硬件特权 + `voiceUnderstanding` 正式模 + 并发闸门）。
- **WS 四模态**：在现有 ASR/TTS 基础上，通过 `start` 参数支持输入 `audio|text`、输出 `audio|text` 四种组合；纯 text 输入时 **仍要求** `sampleRate`/`bits`/`channels` 等音频字段必填（客户端传占位值）。
- **BREAKING**：直接删除外壳 B 及相关 API：
  - `POST /device/history/api/chat`
  - `POST /device/history/api/chat/stream`
  - `POST /voice/internal/api/text/chat`
  - `POST /voice/internal/api/text/chat/stream`
  - `DelegateTextChat` / `DelegateTextChatStream` 及 history 委派实现
  - 调试用 `POST /voice/text/chat`（若仅服务已删路径则一并删除）
- **MCP**：`baby_feeding_advisor` 改为连接 `/voice/chat/ws` 文模式（text 入、text 出），不再依赖 internal text。
- **Admin 喂养 LLM**：保留 lane 键名 `voiceUnderstanding`；删除该 lane 的 `freeProvider`/`freeModel`；Admin 文案可称「喂养默认智能体」，仍可配置 provider/model 与并发。
- **额度 UI**：voice-admin **隐藏** `voice_ai` / 喂养月度额度配置项；对话路径不再消耗 `voice_ai`。
- **设备运维调试台**（`history.html` 或同页区块）：
  - App **用户名密码登录**（`POST /device/app/api/username_login`），与既有 Admin JWT **分钥存放**；登录后用 App Bearer 调用 tip / care-alert。
  - **横屏同款文字对话**：连 `/voice/chat/ws`（text/text），展示 `thinking_delta`/`answer`；不为对话另开鉴权旁路（沿用现网 WS 豁免+Hardware）。
  - **小贴士**：同款 `POST /device/tip/generate`（SSE），展示 thinking/answer；可重复触发（无日缓存；额度按正式 clinic 路径）。
  - **值得留意**：同款 `GET /device/api/care-alert/daily`；支持 **强刷**。
- **care-alert 强刷**：正式 App 接口增加 `force=1`（或等价 query/body）；仍要求 App Bearer 与 `wxId>0`；强制删除当日 Redis 日缓存后重新生成。**不为 tip/care-alert 增加 Admin 鉴权豁免或伪造 `X-Internal-Wx-Id`。**
- **非本变更范围**：Flutter 残留 `sendCommand`/`chat` 调用清理。

## Capabilities

### New Capabilities

- `feeding-chat-ws-io`：`/voice/chat/ws` 四模态 I/O 协议与行为（含 text 帧、输出是否 TTS）。
- `feeding-shell-b-removal`：删除 history/voice internal 文本喂养路径及契约下线要求。
- `feeding-llm-admin-simplify`：`voiceUnderstanding` 去掉 free 字段；额度页隐藏喂养项；硬件通道选模语义。
- `mcp-feeding-ws-text`：MCP 工具改挂 chat WS 文模式。
- `device-ops-ai-debug`：设备运维页 App 登录 + tip/care-alert/WS 文对话调试台。
- `care-alert-force-refresh`：care-alert daily 正式接口强刷（清日缓存后重生）。

### Modified Capabilities

- （基线见 `openspec/specs/v3.0.0/spec.md`）喂养 AI / voiceUnderstanding / history-voice-delegation / chat/stream / MCP / care-alert daily / lane-free-model（喂养侧）等相关 Requirement：本变更以增量 specs 覆盖并在归档合并时 supersede；不在此逐条改 v3 文件。

## Impact

- **进程**：`voice-service`（WS、care-alert force、internal text 删除、VU free）、`history-service`（Chat 删除）、`mcp-service`（WS）、`gateway-app-server`（静态运维页；WS/App API 既有反代）。
- **Admin UI**：`history.html`（调试台+App 登录）、`ai-model-admin.html`、`voice-admin.html`。
- **API**：history/internal text chat **直接下线**；care-alert daily **新增 force 语义**（Additive，非破坏既有无 force 行为）。
- **权益**：喂养对话统一 Hardware；tip/care-alert 仍走正式 App 鉴权与各自额度/VIP。
- **Redis**：复用既有 `CareAlertDailyKey`；force 时 Del 后重生；**不新增**读缓存键族（沿用既有 care-alert 模式）。
- **无新背景 ticker**；无新 DB 库组。

## Context

当前自然语言喂养有两条接入外壳：

1. **外壳 A**：`GET /voice/chat/ws` — PCM→ASR→意图→TTS；`PrivilegeHardware`；网关 Bearer 豁免。
2. **外壳 B**：history `chat(|/stream)` → voice internal text；账号通道可吃 `voice_ai`。

Flutter 已切离外壳 B。设备运维 `history.html` 仅有档案/事件/suggest，无法调试 tip SSE、care-alert 日列表与横屏思考流；care-alert Redis 日缓存使普通重复 GET 无法重生。

约束：文档中文；不新增测试文件；外壳 B **直接删除**；VU 键名不改；tip/care-alert **不做鉴权旁路**；Redis 经 cachekit + 既有 key builder。

## Goals / Non-Goals

**Goals:**

- `/voice/chat/ws` 四模态 I/O。
- 删除外壳 B / internal text / Delegate；MCP 改 WS 文模式。
- VU 去掉 free；额度 UI 隐藏喂养项。
- 运维页：App 用户名登录 + WS 文对话 + tip SSE + care-alert 拉取/强刷。
- care-alert 正式接口 `force` 清日缓存后重生。

**Non-Goals:**

- 清理 Flutter 残留 chat 调用。
- 重命名 `voiceUnderstanding` 键；放宽文模式音频必填。
- 改变 tip/clinic/care-alert 的 VIP∪额度选模策略（除 force 缓存语义外）。
- 为 tip/care-alert 增加 Admin 伪造 wx 头或白名单豁免。
- 新增 Redis 键族或背景 ticker。

## Decisions

### D1：四模态挂在现有 chat WS

- `start` 增加 `inputModality`/`outputModality`（默认 audio/audio）。
- 音频元数据始终必填；文模式传占位值。
- text 帧触发意图；`outputModality=text` 跳过 TTS。

### D2：特权与额度（喂养对话）

- chat WS 继续 Hardware；不计 `voice_ai`；仅 VU 正式模 + 闸门；删除 VU free。

### D3：外壳 B / internal text 直接删除；MCP 改 WS 文模式

### D4：Admin VU / 额度 UI

- VU 保留 provider/model/并发；删 free UI/持久化。
- 额度页隐藏喂养项；API 字段可暂留。

### D5：运维调试台双身份

- **Admin JWT**（既有）：进页、历史 CRUD。
- **App accessToken**（新增，独立 sessionStorage key）：`username_login` 后调用 tip/care-alert。
- 登录成功后 SHOULD 提示/校验返回的 `deviceNo` 与 URL 设备一致，避免串号调试。
- 对话面板：浏览器直连 `/voice/chat/ws`（text/text），不新增鉴权旁路。

### D6：care-alert 强刷用正式 `force`，非 Admin 旁路

- **选择**：`GET /device/api/care-alert/daily?deviceNo=&force=1`（或等价）；校验 App Bearer→`wxId>0` 后，`Del` 当日 `CareAlertDailyKey`（及必要时 lock 键策略与现网一致），再走既有生成/single-flight。
- **备选**：Admin 清缓存 API — 增加第二条运维契约，否决（与「三接口不额外跳过鉴权」一致：强刷仍走 App 身份）。
- 无 `force` 时行为与现网一致（命中日缓存直接返回）。

### D7：tip「强刷」

- tip 服务端无日缓存；运维「强刷」= 再次 POST `/device/tip/generate`（正式鉴权与 clinic 额度照旧）。

### D8：无新库连接；Redis 仅复用 care-alert 既有键

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 已删 chat API 被残留客户端调用 | 产品已切；文档删除路径 |
| 匿名 WS text 刷正式模 | 与硬件通道一致；运维页需知悉 |
| force=1 被客户端滥用重生 | 仍耗 care_alert 额度/VIP 路径；可观测日志打 force |
| 运维 App 账号与设备未绑定 | UI 校验 deviceNo；失败时提示 |
| Admin/App 双 token 混淆 | 分钥；tip/care-alert fetch 只用 App token |

## Migration Plan

1. WS 四模态 → MCP 切 WS → 删外壳 B。
2. VU free 删除 + 额度隐藏。
3. care-alert `force` → history.html 调试台（App 登录 + 三面板）。
4. 更新 runbook / voice-admin 路径说明。

## Open Questions

- MCP WS 基址环境变量名与现网对齐即可。
- `voice_ai` Admin API 字段暂留、仅 UI 隐藏。

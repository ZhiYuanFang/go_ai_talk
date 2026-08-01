## Context

Clinic WS（`/voice/clinic/ws`）在 voice-service 内仍保留三套与「对话/上下文」相关的 Go 侧状态：

1. Redis `voice:clinic:session:{wxId}` + `auth_ok` 后 `session_sync`（给客户端补历史）
2. Redis `voice:clinic:summary:{wxId}:{deviceNo}` + question 前 `DeviceHistory` 聚合 7 天摘要
3. `loadClinicBabyProfile` 每轮拉画像

实际上 `ClinicStream` 只向 Python 发送 `question` / `device_no` / `model`；多轮由 Python `companion_session` 负责，UI 历史由 Flutter 本地 store 负责（且 Flutter 已忽略 `session_sync`）。Go 侧上述路径为死负载，并与「换机/清本地不补对话」冲突。

基线：`openspec/specs/v2.0.24/spec.md` → `pangbao-ai-clinic`。

## Goals / Non-Goals

**Goals:**

- Go Clinic WS 收敛为：鉴权 → 限流/额度/闸门 → Python 流转发 → 下行帧。
- 停止下发 `session_sync`；删除 session / summary Redis 与宝宝画像死路径。
- 同步规格增量与隐私页中「Go 读 7 天摘要」表述。
- 中文注释覆盖关键删除边界（为何不再缓存/同步）。

**Non-Goals:**

- 不改 Flutter / Python 仓代码。
- 不把 clinic 命名批量改为 companion。
- 不改 tip SSE、喂养 voice ball、`clinic_ai` 额度数值与 Admin。
- 不主动 DEL 线上旧 Redis 键（依赖 TTL 自然过期）。
- 不新增测试文件。

## Decisions

### 决策1：协议上直接取消 `session_sync`（非整帧空 turns）

- **方案**：`auth_ok` 写成功后进入读循环；**不下发**任何 `session_sync`。
- **理由**：产品明确「不发」；空帧仍浪费一轮与误导旧文档；现网 Flutter 已 ignore。
- **备选**：恒发空 `turns:[]`——排除，仍保留废弃帧类型。

### 决策2：删除 Go 对话 Redis，不由 Go 补历史

- **方案**：删除 `clinic_session.go` 及 `appendClinicTurn` / `BuildSessionSync` / `prior`；配置去掉 `sessionTtlSeconds`；`cachekit` 去掉 `VoiceClinicSessionKey`（或保留 builder 但无调用方——本变更 **删除 builder**，避免死 API）。
- **理由**：UI=Flutter，agent 多轮=Python；换机/清本地不要求服务端补对话。
- **备选**：保留 Redis 仅供 Python——排除，Python 已有独立键空间。

### 决策3：删除 7 天喂养摘要整条链路

- **方案**：删除 `clinic_summary.go`、`ensureClinicSummary`、`VoiceClinicSummaryKey`、配置 `summaryTtlSeconds`；`HandleQuestion` 不再因摘要失败返回 error。
- **理由**：未注入 Python；每轮还可能触发 history HTTP，纯浪费。
- **备选**：摘要改传 Python body——排除，超出本包且 Python 已可自行拉喂养上下文。

### 决策4：一并删除 `loadClinicBabyProfile`

- **方案**：删除 `clinic_profile.go`；`streamClinicLLM*` 签名去掉 `baby`/`summaryJSON`/`prior`，仅保留 question 流式转发所需参数。
- **理由**：与摘要同属未使用死路径；画像由 Python 侧拉取。
- **备选**：保留降级空画像调用——排除，无消费方。

### 决策5：`deviceNo` 在 Go Clinic 的用途收窄

- **方案**：auth 仍校验 `deviceNo` 与 JWT 一致，并原样传给 Python `device_no`；Go **不再**用它拉 history 摘要或 DeviceProfile。
- **理由**：Python 会话键与上下文依赖 `device_no`；Go 只需透传。

### 决策6：隐私页与配置文案同步

- **方案**：修订 `privacy-policy.html` 中「系统读取近 7 天喂养记录聚合摘要」若归因于本服务 Go 实现的表述，改为不承诺 Go 侧聚合（或改为概括性「AI 能力可能使用喂养相关上下文，由智能服务处理」——实现时与现网隐私政策语气对齐，避免点名已删除的 Go 摘要管线）。
- **方案**：`aiClinic.systemPrompt` 去掉「近 7 天摘要」导向文案（Go 本地拼 LLM 已不存在；若配置块仍保留 systemPrompt 字段但无人读取，可删字段或留空——实现时以「无引用则删配置项」为准）。

### 决策7：Redis 读缓存 / 新键

- **方案**：本变更 **删除** 既有读缓存键用法，**不新增** Redis 键；无需负责人新确认「引入缓存」。限流键 `voice:clinic:rate:*` 保留。

## Risks / Trade-offs

- [旧客户端依赖 session_sync 非空 turns] → 现网 Flutter 已 ignore；文档标明 BREAKING；不发帧即可。
- [Python 未拉到喂养上下文导致回答变「无记录感」] → 属 Python 职责；本包不回填 Go 摘要。
- [隐私页表述与法务口径] → 仅改与已删能力冲突的句子；不扩大隐私承诺范围。
- [删除 cachekit builder 后其他引用] → grep 确认仅 clinic 使用后再删。

## Migration Plan

1. 部署 voice-service 新版本：停止写 session/summary、停止发 session_sync。
2. 旧 Redis 键随 TTL（12h/24h）过期；无需脚本。
3. 回滚：回退二进制即可恢复旧行为（旧键可能已空，session_sync 会变空 turns）。

## Open Questions

- （无）explore 已拍板：不发 session_sync、不补历史、只动 Go、删摘要与 baby profile。

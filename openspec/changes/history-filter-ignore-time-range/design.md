## Context

`GET /device/history/api/filter`（`api/v1.DeviceHistoryFilterReq`）当前用 `startTime`/`endTime` > 0 施加时间条件，=0 跳过。App 与语音/Python 在「时间不确定、只想看之前发生过什么」时仍可能填入猜测区间，导致漏查。本变更在同一路径上增加可选 `ignoreTimeRange`，为真时强制忽略两侧时间参数。

基线：`openspec/specs/v3.0.0/spec.md` 章节 `history-filter-api`。实现落在 `history-service` 及 local/remote/canary 契约链；不新开 v2 路径。

## Goals / Non-Goals

**Goals:**

- 可选 `ignoreTimeRange`，默认 false / 未传 ≡ 现网行为。
- 为真时完全忽略 `startTime`/`endTime`，其余筛选（deviceNo、eventIds、remark、limit、id 倒序）不变。
- App 与语音两侧均可调用；契约 remote HTTP 透传该 query。
- 中文注释覆盖新字段与忽略分支。

**Non-Goals:**

- 不新增 `/v2/filter` 或其它平行路径。
- 不改 `list` / `v2/list` / `piece` 的时间语义。
- 不引入 Redis 读缓存；不改 usage `maintenance_skip`（除非负责人另行要求）。
- 不强制本变更内完成 Python 模型训练/提示词改造（Go 契约与透传就绪即可；若 voice 已有 filter 映射则一并接线）。

## Decisions

### 1. 字段名与极性：`ignoreTimeRange`，默认 false

- **选择**：true = 忽略时间；false = 沿用 start/end 规则。
- **理由**：与「打开开关 = 完全忽略 start/end」一致，避免 `restrictTime` 默认 false 读成「不限制」的歧义。
- **备选**：`restrictTime`（极性易混）；仅约定传 0（无法覆盖「已填错时间」）。

### 2. Additive 扩展现有 v1 路径，不建 v2

- **选择**：在 `DeviceHistoryFilterReq` 增加可选字段；路径/method 不变。
- **理由**：产品明确要求默认 false、不影响原功能；与 care-alert `force` 同类 additive 先例一致。
- **兼容说明**：相对 AGENTS「已有版本结构不可改」的严格字面，本变更记为「可选 query、默认等价旧行为」的兼容例外；归档合并时更新 v3 `history-filter-api`。
- **备选**：新开 v2 filter（更严版本纪律，但调用方双路径成本更高，已否决）。

### 3. 真值解析：bool + 兼容常见 query 字符串

- **选择**：API 类型优先 `bool`（与 `unrepliedOnly` 一致）；若 GoFrame 绑定对 `"1"`/`"true"` 已够用则直接用。文档约定：未传/false/0 → 不忽略；true/1 → 忽略。
- **理由**：GET query 与 LLM 填参常见 0/1；实现阶段以既有绑定行为为准，必要时在 controller 做轻量规范化。
- **备选**：纯 `string` force 风格（如 care-alert）——若 bool 绑定不稳定再退到 string。

### 4. 契约签名扩展

- **选择**：`ListHistoryFilter(..., ignoreTimeRange bool)` 贯穿 contracts / local / remote / switchAdapter；remote 仅在 true 时写 query（或始终写 `ignoreTimeRange=true|false`，以实现简洁为准，推荐 true 才传以减小噪音）。
- **理由**：与现有 `startTime>0` 才写入 query 的风格一致。

### 5. Redis / DB

- **不新增** Redis 读缓存（沿用直查 MySQL + limit 收口；备注探针 limit 上限 20 规则不变）。
- **不新增** DB 连接或表结构变更。

## Risks / Trade-offs

- [忽略时间后扫得更宽] → 仍靠 `limit`（默认 100/上限 500）与备注探针上限 20 收口；不放宽 limit。
- [LLM 误开 ignore 导致结果含窗外数据] → 由调用方/提示词约束；服务端按开关忠实执行。
- [v1 additive 与版本纪律张力] → proposal/design 明示例外；评审按「默认行为不变」验收。
- [语音侧未接线则仅 HTTP 生效] → tasks 区分「契约必做」与「voice 映射若存在则接线」。

## Migration Plan

1. 先部署 history-service（含契约）；旧客户端不传参数 → 行为不变。
2. App / 语音再启用 `ignoreTimeRange=true`。
3. 回滚：回退 history 二进制即可；客户端传未知 query 一般无害。

## Open Questions

- 语音 `IntentEvent` 是否在本变更内对称加字段：若当前 Go 读历史路径尚未调用 `ListHistoryFilter`，可仅预留 HTTP；实现时以代码检索为准。
- usage 是否特殊处理：默认沿用现 filter，**未获负责人答复前不改** `maintenance_skip.go`。

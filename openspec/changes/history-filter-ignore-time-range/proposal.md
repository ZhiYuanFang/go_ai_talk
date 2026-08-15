## Why

`GET /device/history/api/filter` 已支持 `startTime`/`endTime` 为 0 时跳过时间条件，但 App 与语音/Python 在「时间区间不明确、只想查之前发生过什么」时仍常填入猜测时间窗，导致漏结果或空结果。需要显式开关在打开时**强制忽略**已填的 `startTime`/`endTime`，默认关闭以保持现网行为。

## What Changes

- 在现有 `DeviceHistoryFilterReq`（同一 `GET /device/history/api/filter` 路径）上 **additive** 增加可选参数 `ignoreTimeRange`（默认 false / 未传 = 旧行为）。
- 当 `ignoreTimeRange` 为真时，服务端 MUST **完全忽略** `startTime` 与 `endTime`（即使非 0），仅按 `deviceNo`、`eventIds`、`remark`、`limit` 与既有排序返回。
- 契约链同步：controller → `contracts.HistoryService.ListHistoryFilter` → local / remote / canary adapter。
- App 与语音侧均可使用该参数；语音/Python 意图读历史若映射到 filter，应对称支持该开关。
- **非 BREAKING**：默认 false，不传或 false 时行为与现网完全一致。
- **不新增** v2 filter 路径（本变更明确选择在原接口 additive 扩展，作为「默认等价旧行为」的兼容例外；归档合并时更新 v3 基线中 `history-filter-api` 章节）。

## Capabilities

### New Capabilities

- （无）本变更不引入独立新能力域。

### Modified Capabilities

- `history-filter-api`：在既有 filter 筛选 Requirement 上增加 `ignoreTimeRange` 语义与 Scenario（基线见 `openspec/specs/v3.0.0/spec.md` 章节 `history-filter-api`）。

## Impact

- **进程**：`history-service`（实现与 HTTP）；经 gateway/device 反代或 voice→history 契约调用 filter 的路径需透传新 query。
- **API**：`GET /device/history/api/filter` 新增可选 query；路径与鉴权不变；usage 统计沿用现 filter（不改 `maintenance_skip`，除非负责人另行要求）。
- **契约**：`internal/services/contracts`、`history` local/remote/canary、`api/v1/device_history_http.go`、`internal/controller/device_history.go`。
- **语音**：若 `IntentEvent` / Python 读历史映射到 filter，需对称字段；否则仅 HTTP 契约就绪，语音后续单独接线。
- **无**新 DB 连接、无新 Redis 读缓存键、无新背景 ticker。

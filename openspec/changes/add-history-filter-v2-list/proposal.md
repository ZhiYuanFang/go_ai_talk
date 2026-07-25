## Why

Python 服务（python_ai_talk）的 LangGraph 意图分析链路需要按条件筛选历史记录，但 Go 侧 history-service 现有 API 能力不足：
1. `/device/history/api/list` 仅支持分页，不支持按事件ID列表和时间范围筛选，Python 侧传的 startTime/endTime/limit 参数被完全忽略
2. 缺少多事件ID筛选接口，Python 侧只能拉全量数据在应用层过滤，效率低

## What Changes

- 新增 `GET /device/history/api/filter` 接口，支持 deviceNo、eventIds（多ID逗号分隔）、startTime、endTime、limit 参数筛选历史记录
- 新增 `GET /device/history/api/v2/list` 接口，在 v1 分页基础上扩展 startTime、endTime、limit 可选参数（方案A：新建v2，不修改v1）
- v2 接口不传新参数时行为与 v1 完全一致（向后兼容）
- v2 接口传了 limit 时优先使用 limit（替代 pageSize）
- 排序统一使用 id 倒序（与现有 List 接口一致）
- 时间单位统一使用 Unix 秒（与现有 piece 接口一致）

## Capabilities

### New Capabilities

- `history-filter-api`: 新增多条件历史筛选 API，支持事件ID列表 + 时间范围 + limit 筛选
- `history-list-v2-api`: 新增 v2 历史列表 API，支持时间范围和 limit 参数

### Modified Capabilities

<!-- 无：v1 接口保持完全不变，符合接口版本不可修改约束 -->

## Impact

**Go 项目 (go_ai_talk) 修改**:

- `api/v1/device_history_http.go`: 新增 `DeviceHistoryFilterReq` / `DeviceHistoryFilterRes`
- `api/v2/device_history_http.go` (新建): 新增 `DeviceHistoryListReq` / `DeviceHistoryListRes`（v2，带时间范围和 limit）
- `internal/services/contracts/runtime_contracts.go`: `DeviceHistoryContract` 新增 `ListHistoryFilter` 和 `ListHistoryPageV2` 方法
- `internal/services/contracts/http_targets.go`: 新增 `HistoryFilterPath` 和 `HistoryListV2Path`
- `internal/services/history/local.go`: 实现 filter 和 v2 list 查询逻辑
- `internal/services/history/adapter.go`: `localService` / `historyRemoteClient` / `switchAdapter` 三处同步补实现
- `internal/controller/device_history.go`: 新增 `Filter` 和 `ListV2` 方法

**Python 项目 (python_ai_talk)**:

- 本次不修改代码（`get_filtered_history_events` 路径已正确指向 `/device/history/api/filter`，待 Go 侧上线后 `get_history_events` 路径切换到 v2）

**API 影响**:

- 新增 `GET /device/history/api/filter`
- 新增 `GET /device/history/api/v2/list`
- 现有 v1 `GET /device/history/api/list` 完全不变
- 不计入 usage 统计（内部服务调用）
- 无需 Bearer 白名单配置（Python 直连 history-service，不走网关鉴权）

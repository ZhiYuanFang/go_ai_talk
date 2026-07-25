## Context

当前 Go 侧 history-service 提供的历史查询 API 能力有限：

1. **`GET /device/history/api/list` (v1)**：仅支持 `deviceNo/page/pageSize` 分页，忽略 `startTime/endTime/limit` 等可选参数
2. **`GET /device/history/api/piece`**：仅支持单事件ID + 时间区间，用于趋势图，不支持多事件筛选
3. **缺少多事件筛选接口**：Python 侧 `get_filtered_history_events` 已按 `/device/history/api/filter` 路径写好调用代码，但 Go 侧尚无此接口
4. **v1 全量接口能力不足**：Python 侧 `get_history_events` 传了 `startTime/endTime/limit`，但 v1 接口完全忽略

兄弟仓 Python 侧 `app/shared/graphs/nodes/fetch_history.py` 已有完整的降级逻辑：
- filter API 可用时走 filter
- filter 失败时降级到全量 list API
- filter 和 list 都失败时返回空列表

### 约束

- 遵守 AGENTS.md 接口版本不可修改规则：v1 接口完全不动，新增 v2 接口
- 所有新增代码必须有详细中文业务逻辑注释
- 排序统一使用 `id` 倒序（与现有 List 接口一致）
- 时间单位统一使用 Unix 秒（与现有 piece 接口一致）
- 两个接口均为 Python 内部服务调用，不计入 usage 统计，无需 Bearer 白名单
- 历史路由自动命中 `/device/history/api/*` fuzzy 代理，无需额外代理配置

## Goals / Non-Goals

**Goals:**

1. 提供 `GET /device/history/api/filter` 接口，支持多事件ID列表、时间范围、limit 参数筛选历史记录
2. 提供 `GET /device/history/api/v2/list` 接口，在 v1 基础上扩展时间范围和 limit 参数
3. v2 接口不传新参数时行为与 v1 完全一致（向后兼容）
4. 支持 local/remote/canary 三种服务模式下均可正常工作
5. apiregistry 自动加载 v2 路由（走 api/v2/embed）

**Non-Goals:**

1. 不修改 v1 `GET /device/history/api/list` 接口的任何字段和行为
2. 不修改 `/device/history/api/piece` 区段接口
3. 不引入 Redis 缓存（filter 和 v2 list 均直查 DB）
4. 不修改 Python 侧代码（本次只改 Go）
5. 不涉及 CI/CD 流程调整

## Decisions

### D1: filter 接口使用 v1 路径前缀，v2 list 使用 `/v2/list` 路径

**决定**：
- filter：`/device/history/api/filter`（v1 路径，因为是全新接口，不存在版本冲突）
- v2 list：`/device/history/api/v2/list`（显式 v2 前缀，与 v1 区分）

**理由**：
- filter 是全新接口，没有旧版本需要兼容，放在 v1 路径即可
- list 已有 v1，按 AGENTS.md 强制约束必须新建 v2 接口，使用显式 v2 路径前缀
- 两个路径都落在 `/device/history/api/*` fuzzy 代理覆盖范围内，自动走 history 服务切换

**替代方案**：
- filter 也放 `/v2/filter` — 无必要，增加路径复杂度且没有旧版本

### D2: 排序使用 id 倒序而非 startTime 倒序

**决定**：两个新接口统一使用 `ORDER BY id DESC` 排序。

**理由**：
- 与现有 `ListDeviceHistoryPage`（local.go 第121行）保持一致
- `id` 是自增主键，索引性能最好，且天然与写入顺序一致
- startTime 可能存在相同值，排序不稳定
- Python 侧对排序没有特殊要求（只要倒序即最新在前）

### D3: 时间单位使用 Unix 秒

**决定**：`startTime` 和 `endTime` 均使用 Unix 秒（int64），`0` 表示不限制。

**理由**：
- 与现有 `ListHistoryPiece`（piece.go 第23行）接口保持一致
- Python 侧 `time.time()` 返回的就是 Unix 秒
- 数据库 `start_time` 字段本身就是 Unix 秒（BIGINT）
- 避免秒/毫秒换算出错

### D4: v2 list 接口兼容逻辑

**决定**：v2 list 接口的新参数行为：
- `limit > 0` 时：用 `limit` 替代 `pageSize`，`page` 忽略（固定为1），`total` 仍然返回
- `limit <= 0` 时：使用 `page` + `pageSize`，与 v1 行为一致
- `startTime > 0` 时：追加 `WHERE start_time >= startTime`
- `endTime > 0` 时：追加 `WHERE start_time <= endTime`
- `pageSize` 上限保持 100（与 v1 一致），`limit` 上限 500

**理由**：
- Python 侧降级场景只传 `limit=100`，需要能用 limit 拉取足够数据
- 不传新参数时完全等价于 v1，零风险
- total 仍然返回，方便未来前端分页场景复用

### D5: 服务契约扩展策略

**决定**：在 `DeviceHistoryContract` 接口中新增两个独立方法，而不是扩展现有方法签名：

```go
ListHistoryFilter(ctx context.Context, deviceNo string, eventIds []int64, startTime, endTime int64, limit int) ([]entity.History, error)
ListHistoryPageV2(ctx context.Context, deviceNo string, page, pageSize int, startTime, endTime int64, limit int) (HistoryPageResult, error)
```

**理由**：
- 不破坏现有 `ListHistoryPage` 签名，所有调用方不受影响
- local/remote/switchAdapter 三处都需要补实现，独立方法清晰
- remote 模式下可以分别映射到不同的 HTTP 路径

## Risks / Trade-offs

### R1: v2 接口路径 `/v2/list` 可能与未来 v2 版本的其他接口混淆

**风险**：目前只有 list 一个 v2 接口，路径用 `/v2/list`，如果未来有更多 v2 接口，路径模式可能不一致。

**缓解**：保持当前最简实现，未来 v2 接口增多时再统一规划 v2 目录结构。

### R2: limit 设为 500 可能返回较多数据

**风险**：如果用户请求较大时间范围且设备历史数据多，500 条记录可能占用较多内存和网络带宽。

**缓解**：
- 500 条 history 记录（每条约10个字段）数据量不大（约几十 KB）
- Python 侧 judge_data_requirement 中 limit 上限也设为 500（judge_data_requirement.py 第109-111行），两边对齐
- 如后续有性能问题可再下调

### R3: remote/canary 模式下远程 history-service 未升级导致新接口 404

**风险**：如果本地走 remote 模式但远程 history-service 未部署新版本，新接口返回 404。

**缓解**：
- Python 侧已有降级逻辑（filter 失败 → 降级到 list，list 失败 → 返回空）
- Go 侧 switchAdapter 已有 `failoverToLocal` 语义，远程失败可回退本地
- 部署时先升级 history-service 再升级其他服务

### R4: 查询性能

**风险**：多事件ID + 大范围时间查询可能较慢。

**缓解**：
- 有 `device_no` 过滤（基数大），加上 `id` 倒序索引，查询性能应该可接受
- `limit` 限制返回条数，避免全表扫描
- 如后续发现性能问题，可考虑加 `(device_no, start_time)` 复合索引

## Migration Plan

### 部署步骤

1. 部署升级后的 history-service（包含 filter 和 v2 list 接口）
2. Python 侧无需改动（`get_filtered_history_events` 路径已正确）
3. 后续 Python 侧可将 `get_history_events` 路径切换到 `/device/history/api/v2/list`（非必须，v1 仍然可用）

### 回滚策略

- v1 接口完全未动，新接口失败不影响任何现有功能
- 如需回滚：只需部署旧版本 history-service，新接口返回 404，Python 侧自动降级

## Open Questions

无（所有设计决策已在探索阶段确认）。

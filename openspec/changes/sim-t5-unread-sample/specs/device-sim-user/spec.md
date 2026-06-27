## ADDED Requirements

### Requirement: device internal SHALL provide full simulated wx id list

device-service MUST 提供 `GET /device/internal/api/sim/wx/ids`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在单条有界 SQL 内返回 **全部** `is_simulated=1` 的 wxId 列表，响应 `{ ids: int64[], total: int }`。MUST NOT 分页截断（不得仅返回前 200 条）。当 `total` 超过 **10000** 时 MUST 返回 4xx 且 MUST NOT 返回部分 ids。

#### Scenario: Ids returns all simulated users

- **WHEN** 库中存在 350 条 `is_simulated=1` 且请求携带有效密钥
- **THEN** 响应 `total` MUST 为 350 且 `ids` 长度 MUST 为 350

#### Scenario: Ids empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 为 `{ ids: [], total: 0 }` 且 MUST NOT 500

#### Scenario: Over limit rejected

- **WHEN** `is_simulated=1` 计数超过 10000
- **THEN** MUST 返回错误且 MUST NOT 返回 ids 数组

## MODIFIED Requirements

### Requirement: device internal SHALL list and batch-query simulated users

device-service MUST 提供：

- `GET /device/internal/api/sim/wx/list` — 分页返回 `is_simulated=1` 的 wxId 与 account 列表（供 Admin、计数 total 等列举场景）
- `GET /device/internal/api/sim/wx/random` — 返回单条随机 simulated 用户（供 sim 任务随机选取）
- `GET /device/internal/api/sim/wx/ids` — 返回**全量** simulated wxId（供 T5 未读抽样等 MUST 覆盖全库 sim 的场景）
- sim 批量查询 MUST 在现有 `POST /device/internal/api/ucg/wx/batch` 响应项中增加 `isSimulated` 字段，或提供等价的 sim 专用 batch 接口

#### Scenario: List sim users

- **WHEN** sim-service 请求 sim wx list 分页
- **THEN** 返回列表 MUST 仅含 `is_simulated=1` 的 wxId

#### Scenario: Batch includes flag

- **WHEN** UCG 或 gateway 批量查询 wxId 展示字段
- **THEN** 每项 MUST 含 `isSimulated` 布尔值

#### Scenario: List not required for random pick

- **WHEN** sim-service 需随机选取一个模拟用户执行任务
- **THEN** MUST 使用 random 接口，MUST NOT 依赖 list 分页结果在客户端 `rand` 选 wxId

#### Scenario: T5 must not use list first page as full sim set

- **WHEN** sim-service T5 需要全库 sim wxId 集合
- **THEN** MUST 使用 ids 接口，MUST NOT 使用 `sim/wx/list?page=1&pageSize=200` 代替全集

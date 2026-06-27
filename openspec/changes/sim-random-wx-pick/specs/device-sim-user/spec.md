## ADDED Requirements

### Requirement: device internal SHALL provide random single simulated wx pick

device-service MUST 提供 `GET /device/internal/api/sim/wx/random`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在 `is_simulated=1` 集合上通过有界 ID 探测返回 **0 或 1** 条 `{wxId, account}`，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 覆盖全库 simulated 用户（非仅第一页）。锚点 MUST 在 `[minId, maxId]` 上 **均匀** 生成：`R = minId + floor((maxId - minId) * U)`（`U` 为 `(0,1)` 均匀随机），随后 `WHERE is_simulated=1 AND id >= R ORDER BY id ASC LIMIT 1`；锚点落空且 eligible 存在时 MUST 回退 `minId` 一条。MUST NOT 对 high-id / 新注册用户做幂次偏置。

#### Scenario: Random returns one sim user

- **WHEN** sim-service 携带有效密钥 GET random 且存在至少一条 `is_simulated=1`
- **THEN** 响应 MUST 含 `wxId>0` 与非空 `account`

#### Scenario: Random empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 表示无结果（空或 found=false），且 MUST NOT 500

#### Scenario: Bounded SQL only

- **WHEN** 代码评审 random 实现
- **THEN** MUST 为 MIN/MAX 聚合 + `LIMIT 1` 探测，MUST NOT 全表加载或 `ORDER BY RAND()`

## MODIFIED Requirements

### Requirement: device internal SHALL list and batch-query simulated users

device-service MUST 提供：

- `GET /device/internal/api/sim/wx/list` — 分页返回 `is_simulated=1` 的 wxId 与 account 列表（供 Admin、计数 total 等列举场景）
- `GET /device/internal/api/sim/wx/random` — 返回单条随机 simulated 用户（供 sim 任务随机选取，见 ADDED Requirement）
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

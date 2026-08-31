## Why

`prediction_unlock.totalActivatableCount` 现按「至少有一个子节点的事件」统计，排除了一级无子根（如「洗澡」）。产品预测可激活槽位应对齐**事件字典一级根**（含无子根），否则天花板偏小（例如库有 8 个一级根却只返回 4）。

## What Changes

- **BREAKING（语义）**：device 供 cash 的可激活事件计数从「非叶子（有子节点）」改为**一级根**（`parent_id = 0`），含无子根。
- cash catalog 仍写 `totalActivatableCount`，数值改为一级根个数；失败时仍不得返回误导性正数。
- 注释 / 内部 API 文案与规格改称「一级根 / root」，避免与 voice「非叶子须追问」术语冲突；HTTP 路径可暂保留 `non-leaf-count` 以免无谓破坏内部调用方（语义以响应 count 为准），或同步改名（实现期二选一，默认保留路径改语义）。
- 客户端预测锁槽位 MUST 与「一级根」同一集合（若 App 仍按旧非叶子列表对齐，须同步）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `catalog-activatable-total`：计数口径改为事件字典一级根数量；catalog `totalActivatableCount` 随之变更。

## Impact

- **device-service**：`CountNonLeafEvents`（或重命名）与 `GET /device/internal/api/event/non-leaf-count`
- **cash-service**：`FetchNonLeafEventCount` / catalog 聚合注释；响应字段名不变
- **Flutter（孪生）**：预测「全部激活」与锁下标须按一级根；本仓无 Flutter 代码
- **非目标**：改 voice 追问非叶子定义；改 `allowedCount` / VIP；新建测试文件

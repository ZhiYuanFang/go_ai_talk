## ADDED Requirements

### Requirement: sample API SHALL support random mode via ID probe with mild recency bias

当请求 body `mode` 为 `random` 时，系统 MUST 在 ucg 库内通过有界 ID 探测返回 **0 或 1** 条 `status=published` 帖子，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 覆盖全库 published 帖（非仅最新 N 条）。锚点 MUST 在 `[minId, maxId]`（published 帖的 id 最小/最大值）上按 `R = minId + floor((maxId - minId) * U^α)` 生成，其中 `U` 为 `(0,1)` 均匀随机数，默认 `α = 0.65`。随后 MUST 执行 `WHERE status=published AND id >= R ORDER BY id ASC LIMIT 1` 取帖。响应字段 MUST 与 latest 模式一致（`postId`、`content`、`mediaType`、可选 `coverObjectKey`）。`mode=random` 时 MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 device HTTP。

#### Scenario: Random mode returns one published post

- **WHEN** 请求携带有效 internal 密钥且 body 为 `{ "mode": "random" }`，且存在至少一条 published 帖
- **THEN** 响应 `list` MUST 含 1 条帖子，且该帖 `status` MUST 为 published

#### Scenario: Random mode empty plaza

- **WHEN** 请求 `mode=random` 且无 published 帖
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Random mode does not use RAND sort

- **WHEN** 代码评审 random 模式 SQL
- **THEN** MUST NOT 出现 `ORDER BY RAND()`

#### Scenario: Latest mode unchanged

- **WHEN** 请求未传 `mode` 或 `mode=latest`，且 `limit` 为 20
- **THEN** 响应 MUST 仍按 `published_at DESC` 返回最多 20 条（与变更前一致）

## MODIFIED Requirements

### Requirement: sample API MUST use bounded single-query read pattern

抽样读路径 MUST 在 ucg 库内完成，MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 `Device().BatchWx`。

- **latest 模式**（缺省）：单条有界 SQL，`LIMIT` ≤ 50；`limit` 默认 20，超出 MUST 截断为 50；排序 `published_at DESC`。
- **random 模式**：MUST 使用有界读模式：一次 `MIN(id)/MAX(id)` 聚合（`WHERE status=published`）加一次 `id >= R LIMIT 1` 探测（共 2 次 SQL）；每次 MUST 带 `LIMIT 1` 或聚合有界，MUST NOT 全表加载。

#### Scenario: Limit clamp in latest mode

- **WHEN** 请求 `limit` 为 100 且未指定 random 模式
- **THEN** 实际查询 MUST 最多返回 50 条

#### Scenario: Random mode bounded queries

- **WHEN** 请求 `mode=random`
- **THEN** 读路径 MUST 最多 2 次 SQL且最终 MUST NOT 返回超过 1 条

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审 ucg internal sample 实现（含 random 分支）
- **THEN** MUST NOT import device 域 DAO 或直连 device 库表

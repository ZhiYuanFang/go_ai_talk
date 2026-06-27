## MODIFIED Requirements

### Requirement: sample API MUST use bounded single-query read pattern

抽样读路径 MUST 在 ucg 库内完成，MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 `Device().BatchWx`。

- **latest 模式**（缺省）：单条有界 SQL，`LIMIT` ≤ 50；`limit` 默认 20，超出 MUST 截断为 50；排序 `published_at DESC`。
- **random 模式**：MUST 使用有界读模式：一次 `MIN(id)/MAX(id)` 聚合（`WHERE status=published` 及可选 author 排除）加一次 `id >= R LIMIT 1` 探测（共 2 次 SQL）；每次 MUST 带 `LIMIT 1` 或聚合有界，MUST NOT 全表加载。
- 当请求 body 含非空 `excludeAuthorWxIds` 时，latest 与 random 模式的 published 帖查询 MUST 附加 `author_wx_id NOT IN (...)`（与 T5 `simWxIds` 排除 sim peer 语义对称）。

#### Scenario: Limit clamp in latest mode

- **WHEN** 请求 `limit` 为 100 且未指定 random 模式
- **THEN** 实际查询 MUST 最多返回 50 条

#### Scenario: Random mode bounded queries

- **WHEN** 请求 `mode=random`
- **THEN** 读路径 MUST 最多 2 次 SQL且最终 MUST NOT 返回超过 1 条

#### Scenario: Exclude sim authors in random mode

- **WHEN** 请求 `mode=random` 且 `excludeAuthorWxIds` 含全部 sim wxId，且存在至少一条 published 帖其 author 不在该集合
- **THEN** 响应 `list[0].authorWxId` MUST 不在 `excludeAuthorWxIds` 中

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审 ucg internal sample 实现（含 random 分支）
- **THEN** MUST NOT import device 域 DAO 或直连 device 库表

### Requirement: sample API response SHALL include authorWxId for internal consumers

`POST /ucg/internal/api/posts/sample` 响应 `list` 每项 MUST 含 `authorWxId`（`ucg_post.author_wx_id`）。T2 等既有调用方 MAY 忽略该字段。

#### Scenario: Author field present

- **WHEN** sample 返回非空 `list`
- **THEN** 每条 MUST 含 `authorWxId` 且 MUST 大于 0

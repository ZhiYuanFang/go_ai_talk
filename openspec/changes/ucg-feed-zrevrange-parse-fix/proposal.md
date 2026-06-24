## Why

测试栈 MONITOR 证实：Feed 读路径 `ZREVRANGE ucg:recommend:score WITHSCORES` 返回数据正常，但 `cachekit.SortedSetRevRangeWithScores` 按**扁平** `[m,s,m,s]` 解析 go-redis 9.x 的**嵌套** `[[m,s],...]` 响应，导致 `z.Member` 变为 `["1",1.718...]` 字符串、`ParseUint` 全失败、候选池恒空。lazy warm / no-geo fix 均已生效且 Redis 索引齐全时 Feed 仍 `list=[]`，根因在本解析层而非索引或半径语义。

## What Changes

- 修正 `internal/platform/cachekit` 中 `SortedSetRevRangeWithScores`（及必要时共用 helper），兼容 go-redis / GoFrame `Do(ZREVRANGE ... WITHSCORES)` 的嵌套与扁平两种数组形态。
- 解析后 `member` MUST 为 Redis ZSET member 字符串（如 `"1"`），`score` 为对应 float；**MUST NOT** 将整对 `[member,score]` JSON 当作 member。
- runbook 补充一条排查：`GEOPOS` 参数若出现 `["id",score]` 形态 → 怀疑 ZREVRANGE 解析 bug。
- **无** Redis schema / HTTP API / usage 统计变更；**无** Flutter 必改项。

## Capabilities

### New Capabilities

- `cachekit-zrevrange-parse`：`cachekit` 对 Redis `ZREVRANGE ... WITHSCORES` 结果的 member/score 解析契约（嵌套/扁平双形态）。

### Modified Capabilities

- `ucg-recommend-feed`：当 `ucg:recommend:score` ZCARD>0 且 snapshot 齐全时，Feed **MUST** 能返回非空 list（与既有 composite 读路径规格对齐；本 change 修复阻碍该行为的 platform 解析缺陷）。

## Impact

- **go_ai_talk**：`internal/platform/cachekit/cache_geo.go`（主修复）；可选抽取 `zset_parse.go` 与 `hash_parse.go` 同族 helper。
- **ucg-service**：Feed / reconciler / 凡调用 `SortedSetRevRangeWithScores` 的路径均受益，无需业务层改动。
- **与 lazy warm / no-geo fallback 关系**：索引与半径语义已 OK；本 change 修复 ZSET 候选**读不出 member** 的 platform bug。

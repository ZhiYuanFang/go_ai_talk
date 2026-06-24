## ADDED Requirements

### Requirement: ZREVRANGE WITHSCORES 解析兼容嵌套与扁平响应

`cachekit` 的 `SortedSetRevRangeWithScores`（及内部 helper）在解析 Redis `ZREVRANGE key start stop WITHSCORES` 结果时，MUST 同时支持：

- **嵌套形态**：`[[member, score], [member, score], ...]`（go-redis 9.x 经 GoFrame `Do` 常见返回）；
- **扁平形态**：`[member, score, member, score, ...]`（历史/其他 adapter 返回）。

解析输出的每个 `ZSetMemberScore.Member` MUST 为 Redis ZSET 的 member 字符串（如 `"1"`），MUST NOT 为整对 `[member,score]` 的 JSON 或数组字符串。`Score` MUST 为对应浮点分值。

#### Scenario: 嵌套响应解析为正确 member

- **WHEN** `Do(ZREVRANGE ... WITHSCORES)` 返回 `[[ "1", "1.718" ], [ "19", "1.195" ]]`
- **THEN** `SortedSetRevRangeWithScores` MUST 返回 `[{Member:"1", Score:1.718}, {Member:"19", Score:1.195}]`（score 允许浮点误差）

#### Scenario: 扁平响应仍兼容

- **WHEN** 响应为 `[ "1", "1.718", "19", "1.195" ]`
- **THEN** 解析结果 MUST 与嵌套形态等价

#### Scenario: 空 ZSET

- **WHEN** Redis 返回空数组或 nil
- **THEN** MUST 返回空 slice 且无 error

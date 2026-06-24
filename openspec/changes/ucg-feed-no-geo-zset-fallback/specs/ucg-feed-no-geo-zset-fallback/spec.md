## ADDED Requirements

### Requirement: viewer 有坐标时 unlimited 半径步 SHALL 从 ZSET 补全无 GEO 帖

当 viewer 请求含有效 lat/lng 且 GEO 半径阶梯已执行至 `radiusKm=0`（unlimited）时，Feed 候选收集 MUST 从 `ucg:recommend:score` ZSET 分页扫描成员并加入候选集，**MUST NOT** 在 unlimited 步仍仅使用 GEOSEARCH。ZSET 中 **不在** `ucg:feed:geo` 的 published 帖（无坐标）MUST 可进入 Feed，按 `baseScore` 参与 `finalScore`（`distanceTerm=0`），响应 MUST NOT 含 `distanceMeters`。

已在较小 GEO 半径（50–500km）或 pool 中的帖 MUST 由 pool/session 去重，不得重复下发。

#### Scenario: 帖仅在 ZSET 不在 GEO，viewer 带坐标

- **WHEN** published 帖在 `ucg:recommend:score` 有 member 但 `GEOPOS ucg:feed:geo` 为空，且 `GET /feed/recommend` 首屏含有效 lat/lng
- **THEN** 响应 `list` MUST 含该帖（在 pageSize 与候选 batch 覆盖范围内）

#### Scenario: unlimited 步不使用 GEO 替代 ZSET

- **WHEN** viewer 带坐标且当前半径阶梯为 `radiusKm=0`
- **THEN** 候选收集 MUST 使用 ZSET 读路径，MUST NOT 以 GEO 20000km 搜索作为 unlimited 步的唯一来源

#### Scenario: 无 viewer 坐标行为不变

- **WHEN** `GET /feed/recommend` 不含有效 lat/lng
- **THEN** 候选收集 MUST 仍按 ZSET/baseScore 排序，与变更前一致

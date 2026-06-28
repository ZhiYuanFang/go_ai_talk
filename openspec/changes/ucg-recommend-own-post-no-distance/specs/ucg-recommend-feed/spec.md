## ADDED Requirements

### Requirement: 推荐 Feed 对作者本人帖 SHALL omit distanceMeters

当已登录 viewer 请求 `GET /ucg/app/api/feed/recommend` 且请求含有效 `lat`/`lng`、帖子 snapshot 含坐标时，若帖子 `authorWxId` 等于 viewer 的 `wxId`，该 item **MUST NOT** 含 JSON 字段 `distanceMeters`（即使 viewer 与帖坐标可计算 haversine）。本要求 **MUST NOT** 改变该帖的 `finalScore` 或 composite 排序语义。

#### Scenario: 本人帖 omit 距离

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且列表含其本人发布的 published 帖（帖含坐标）
- **THEN** 该 item JSON MUST NOT 含 `distanceMeters`
- **AND** 同页他人帖在 viewer 与帖均有坐标时 MUST 仍含 `distanceMeters`

#### Scenario: 本人帖排序不变

- **WHEN** viewer 本人帖与他人帖参与同一推荐页 composite 排序
- **THEN** 系统 MUST 仍按既有 `finalScore = baseScore + distanceTerm` 排序
- **AND** omit `distanceMeters` MUST NOT 单独改变该帖的 `finalScore` 计算

#### Scenario: 未登录无本人概念

- **WHEN** 匿名或未带 `X-Internal-Wx-Id` 的有效登录上下文请求推荐 Feed
- **THEN** 距离字段行为 MUST 与变更前一致（有坐标则按他人帖规则返回 `distanceMeters`）

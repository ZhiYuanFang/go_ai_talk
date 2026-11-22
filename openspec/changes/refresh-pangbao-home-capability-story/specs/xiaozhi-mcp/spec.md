## MODIFIED Requirements

### Requirement: App SHALL 返回已配置引导视频的 CDN 列表

系统 MUST 提供 `GET /cash/app/api/xiaozhi-mcp/guide-videos`（经 gateway-app 暴露）。 响应 `list` MUST 仅含 `object_key` 非空的槽，每项至少含 `slot`、`title`、`videoCdnUrl`； 该接口 MUST **允许匿名访问**（gateway-app Bearer 豁免或等价公开路径），以便官网未登录拉取； 携带有效 App Bearer 时行为 MUST 与匿名成功响应一致（同 list 语义）。 是否计入 usage 统计 MUST 以负责人确认结果为准（实现前询问； 未确认前 MUST NOT 擅自加入或移出 `maintenance_skip`）。 若采用独立公开路径（如 `/cash/app/api/xiaozhi-mcp/guide-videos/public`），则该路径 MUST 与主路径返回同一 list 语义，且 MUST 豁免 Bearer； 主路径 MAY 继续要求 Bearer（但官网 MUST 使用已豁免路径）。

#### Scenario: 仅一槽有视频

- **WHEN** 仅 `promo` 已配置视频
- **THEN** `list` 长度 MUST 为 1 且含 `slot=promo`

#### Scenario: 匿名可读

- **WHEN** 未登录客户端请求已豁免的引导视频列表 GET
- **THEN** 系统 MUST 返回成功且 `list` 仅含已配置槽（MUST NOT 因缺少 Bearer 拒绝）

#### Scenario: 已登录 App 仍可读

- **WHEN** App 携带有效 Bearer 请求引导视频列表
- **THEN** 系统 MUST 返回与匿名成功时相同语义的已配置槽列表

#### Scenario: 未登录拒绝

- **WHEN** 请求无有效 App Bearer
- **THEN** MUST 拒绝（MUST NOT 匿名放行）


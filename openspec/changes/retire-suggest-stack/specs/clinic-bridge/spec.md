# Spec Delta

## MODIFIED Requirements

### Requirement: 聊天成长建议 MUST 删除

系统 MUST 删除喂养聊天中的「成长建议」生成路径（含 `callDeepSeekGrowthSuggestion` 及仅为其服务的 payload/意图分支），MUST NOT 再调用 Python `POST /v1/clinic` 同步陪伴接口。系统 MUST 拆除 Hub/history/voice 上剩余的 suggest 列表与删除运维面及对应 HTTP API；MUST NOT 再向运维或 App 暴露成长建议读删入口。历史 `suggest` 表行 MAY 物理留库，但 MUST NOT 再作为产品或运维功能表面。

#### Scenario: 不再触发成长建议 LLM

- **WHEN** 用户走喂养语音聊天主路径
- **THEN** 系统 MUST NOT 调用成长建议专用生成函数，MUST NOT 请求 `/v1/clinic`

#### Scenario: 运维与 App 无 suggest 读删口

- **WHEN** 调用方请求原成长建议列表或删除路径（含 Hub 历史页聚合口与 voice 内部 suggest API）
- **THEN** 系统 MUST NOT 再提供可成功返回业务列表或执行删除的正式路由

## ADDED Requirements

### Requirement: 成长建议产品面 MUST 保持退役

在生成路径已删除的前提下，系统 MUST 保持成长建议无写入、无运维列表、无 App/Hub 可读删契约；后续变更 MUST NOT 在未新开 OpenSpec 变更的情况下恢复 suggest 业务 API。

#### Scenario: 无新写入来源

- **WHEN** 任一喂养或 clinic 相关上游完成一轮对话
- **THEN** 系统 MUST NOT 向 `suggest` 表插入新业务行（遗留表可存在但无产品写入路径）

## ADDED Requirements

### Requirement: 门户按 document 查询九个飞轮

voice-service MUST 提供需管理员身份的 `GET /voice/admin/api/intent-cues/by-document`。它 MUST 把查询参数 `document` 转发给 Python `GET /v1/admin/cues/by-document`，并返回该接口的命中列表。去除空白后的 `document` 为空时，服务端 MUST 拒绝且 MUST NOT 调用 Python。该接口 MUST NOT 通过逐页拉取各维度列表再在本地过滤来实现。

#### Scenario: 转发精确查询

- **WHEN** 管理员传入一句非空原文
- **THEN** voice-service MUST 向 Python 聚合接口查询该原文，并把命中的维度、id 与 payload 返回给 Hub

### Requirement: 命中行可按维度删除

意图智能体门户 MUST 提供原文输入与查询。每条命中 MUST 单独占一行，并展示维度、向量 id 与 payload。对已知维度，删除 MUST 调用该维度已有的 `DELETE /voice/admin/api/*-vectors/{id}`，且 MUST NOT 用同一个删除路径删除另一个维度的 id。

#### Scenario: 删除备注命中

- **WHEN** 管理员删除一条 `dimension=remark` 的命中
- **THEN** 门户 MUST 调用 `DELETE /voice/admin/api/remark-vectors/{id}`，其中 id 为该命中的向量 id

#### Scenario: 空输入不查询

- **WHEN** 管理员在原文为空时点击查询
- **THEN** 门户 MUST NOT 发起查询请求

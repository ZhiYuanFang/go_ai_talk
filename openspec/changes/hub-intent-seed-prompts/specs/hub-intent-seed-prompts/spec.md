## ADDED Requirements

### Requirement: 各维向量页提供可复制的种子提示词

意图门户所链的 OP、事件、数量、时间、备注、R·数据源、R·predict、查史播报、U/D 定位管理页，MUST 在批量上传区展示只读提示词，并 MUST 提供复制到剪贴板的操作。提示词 MUST 要求其它模型只输出 JSON 数组，元素含 `document` 与 `payload`，且 payload 字段与枚举与当前 Python 该集合约定一致。除事件维（短说法，payload 为 `event_id`/`event_name`/`kind`）外，document MUST 说明为整段原话。

#### Scenario: 复制预测维提示词含三字段

- **WHEN** 运维打开 R·predict 向量页并复制种子提示词
- **THEN** 文本包含 `predict_channel`、`predict_aspect`、`time_answer` 的合法取值说明
- **AND** 同页批量示例 JSON 含 `predict_aspect`

#### Scenario: 时间与定位禁止项写在提示词内

- **WHEN** 运维复制时间维或定位维提示词
- **THEN** 时间维文本禁止以绝对 Unix 作为主缓存
- **AND** 定位维文本禁止 `historyId`

### Requirement: 类型向量不回到门户

`feeding_type_cues` 管理页 MUST NOT 重新出现在意图智能体门户的飞轮入口中。若该页展示提示词，文案 MUST 标明热路径已停。

#### Scenario: 门户无类型向量卡片

- **WHEN** 运维打开意图智能体门户
- **THEN** 飞轮入口不包含类型向量管理页链接

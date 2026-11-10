## ADDED Requirements

### Requirement: 事件字典 MUST 支持合并预测标志且默认否

系统 MUST 在全局事件字典（`event`）持久化合并预测标志：数据库列名为 `merge_predict`，对外 JSON 字段名为 `mergePredict`。新建事件时若未显式指定，该标志 MUST 为 false（0）。该标志 MUST NOT 复用或改写既有 `eventType`、`isAppointment` 或 `parent_id` 语义。

#### Scenario: 新建事件默认不合并预测

- **WHEN** 管理员新增事件且未勾选合并预测
- **THEN** 持久化记录的 `merge_predict` MUST 为 0，且后续列表/options 返回中该事件 `mergePredict` 为 false/0

#### Scenario: 与预约标志正交

- **WHEN** 某根事件同时存在 `isAppointment` 与 `mergePredict` 取值
- **THEN** 系统 MUST 独立持久化与暴露两字段，MUST NOT 因其一为真而改写另一字段

### Requirement: 仅有子节点的根事件 MAY 将合并预测置为真

系统 MUST 仅在事件同时满足以下条件时允许将 `merge_predict` 持久化为 1：（1）`parent_id` 为 0（根事件）；（2）存在至少一条 `parent_id` 等于该事件 id 的子行。不满足条件时，写路径 MUST 将标志写为 0（即使请求体传 1）。

#### Scenario: 有子的根可开启合并预测

- **WHEN** 管理员对 `parent_id=0` 且已有直接子节点的事件更新 `mergePredict=1` 且保存成功
- **THEN** 该行 `merge_predict` MUST 为 1，且后续 options/list 返回 `mergePredict` 为真

#### Scenario: 无子的根不可开启

- **WHEN** 管理员对 `parent_id=0` 但无任何子行的事件请求 `mergePredict=1`
- **THEN** 持久化结果 MUST 为 `merge_predict=0`

#### Scenario: 叶子事件不可开启

- **WHEN** 管理员对 `parent_id≠0` 的事件请求 `mergePredict=1`
- **THEN** 持久化结果 MUST 为 `merge_predict=0`

#### Scenario: 根改挂为非根时清零

- **WHEN** 某根事件原 `merge_predict=1` 且更新将其 `parent_id` 改为非 0
- **THEN** 更新后该行 `merge_predict` MUST 为 0

### Requirement: 删除导致父变为无子根时 MUST 清零父的合并预测

当删除事件成功后，若被删事件的原父节点变为「根且无子」，系统 MUST 将该父节点的 `merge_predict` 置为 0，并触发与既有事件变更一致的 EventOptions 缓存重建。

#### Scenario: 删光最后子节点清零父标志

- **WHEN** 根事件 R 仅有一子 C 且 R 的 `merge_predict=1`，管理员成功删除 C
- **THEN** R 的 `merge_predict` MUST 变为 0

### Requirement: 后台 MUST 可按规则编辑合并预测标志

设备域 Hub 事件新增/编辑表单 MUST 提供合并预测勾选；仅当当前编辑对象为有子节点的根时勾选控件可用，否则 MUST 禁用。更新成功后，系统 MUST 使事件字典读模型与既有「变更后重建 EventOptions 缓存」约定一致。

#### Scenario: Hub 对无子根禁用勾选

- **WHEN** 运维打开无子根事件的编辑表单
- **THEN** 合并预测勾选控件 MUST 为禁用状态

#### Scenario: Hub 开启后可读

- **WHEN** 运维对有子根成功勾选合并预测并保存
- **THEN** 之后经事件列表或 options 读取该事件时 MUST 返回 `mergePredict` 为真

### Requirement: 事件 options 与列表 MUST 附带合并预测标志

凡返回 `entity.Event`（或等价事件字典项）的 App/内部 options 与后台列表路径，MUST 包含 `mergePredict` 字段，供 Flutter 预测算法读取。字段缺失 MUST 被客户端视为 false；实现 MUST 保证新列映射到响应（旧缓存淘汰后）。

#### Scenario: Flutter 可从 options 识别合并预测根

- **WHEN** 客户端请求历史事件 options（或等价字典接口）且字典中存在已标记合并预测的根事件
- **THEN** 响应列表中对应项 MUST 包含 `mergePredict` 为真

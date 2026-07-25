## ADDED Requirements

### Requirement: Python 返回完整的 CRUD 所需数据

Python 意图分析服务 SHALL 返回可直接用于 CRUD 的完整数据，包括：
- `event_id`：事件唯一标识（已有事件时返回）
- `event_name`：事件名称
- `quantity`：从用户输入中提取的数量值
- `event_type`：事件类型（新事件时返回，值为 number/time/one）
- `event_unit`：事件单位（新事件时返回，如 ml、次、分钟）
- `is_new_event`：是否为新事件

#### Scenario: 已有事件且能提取数量

- **WHEN** 用户输入"刚才喝了120ml奶粉"
- **AND** Python 向量匹配置信度 > 0.95
- **AND** 能提取数量值
- **THEN** Python 返回：
  ```json
  {
    "target_type": "feeding",
    "action": "one",
    "event_id": "123",
    "event_name": "喂奶",
    "quantity": 120,
    "is_new_event": false
  }
  ```

#### Scenario: 已有事件但无法提取数量

- **WHEN** 用户输入"喂了奶"
- **AND** Python 向量匹配置信度 > 0.95
- **AND** 无法提取数量值
- **THEN** Python 返回：
  ```json
  {
    "target_type": "feeding",
    "action": "one",
    "event_id": "123",
    "event_name": "喂奶",
    "quantity": null,
    "is_new_event": false
  }
  ```

#### Scenario: 新事件需要推断类型和单位

- **WHEN** 用户输入"游泳了30分钟"
- **AND** 向量匹配无法找到已有事件
- **THEN** Python 返回：
  ```json
  {
    "target_type": "feeding",
    "action": "one",
    "event_name": "游泳",
    "quantity": 30,
    "event_type": "time",
    "event_unit": "分钟",
    "is_new_event": true
  }
  ```

### Requirement: 前置数量提取

Python 意图分析 SHALL 在向量匹配后、LLM 分类前，使用正则表达式提取数量：
- 支持汉字数字转换（一→1、二→2、...、九→9）
- 支持阿拉伯数字提取（如 120、90）
- 高置信度匹配（>0.95）且能提取数量时，跳过 LLM 调用直接返回

#### Scenario: 汉字数字转换

- **WHEN** 用户输入"喝了三瓶奶"
- **THEN** Python 提取数量为 3

#### Scenario: 阿拉伯数字提取

- **WHEN** 用户输入"刚才喝了120ml"
- **THEN** Python 提取数量为 120

#### Scenario: 无数字时返回 null

- **WHEN** 用户输入"喂了奶"
- **THEN** Python 返回 quantity 为 null

### Requirement: Go 侧根据 event_id 查询 Redis

Go 侧 SHALL 根据 Python 返回的 `event_id` 查询 Redis 缓存获取事件的 `event_type` 和 `event_unit`。

#### Scenario: Redis 缓存命中

- **WHEN** Python 返回 event_id = "123"
- **AND** Redis 缓存中存在该事件
- **THEN** Go 侧从 Redis 获取 event_type 和 event_unit

#### Scenario: Redis 缓存未命中回源 MySQL

- **WHEN** Python 返回 event_id = "123"
- **AND** Redis 缓存中不存在该事件
- **THEN** Go 侧查询 MySQL 获取事件信息
- **AND** 将结果写入 Redis 缓存

### Requirement: Go 侧处理新事件入库

当 Python 返回 `is_new_event: true` 时，Go 侧 SHALL 使用 Python 返回的 `event_type` 和 `event_unit` 创建事件，并更新 Redis 缓存。

#### Scenario: 新事件入库成功

- **WHEN** Python 返回：
  ```json
  {
    "event_name": "游泳",
    "event_type": "time",
    "event_unit": "分钟",
    "is_new_event": true
  }
  ```
- **THEN** Go 侧创建新事件，入库成功后更新 Redis 缓存
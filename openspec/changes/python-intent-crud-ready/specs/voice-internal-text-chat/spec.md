## MODIFIED Requirements

### Requirement: Go 侧删除数量提取和自然语言匹配逻辑

Go 侧语音对话服务 SHALL 删除以下逻辑：
- `extractNumberFromText` 函数及其调用
- `extractEventFromCandidates`、`hasSignificantOverlap`、`sortForMatch` 等自然语言匹配函数
- `callDeepSeekEntityExtract` DeepSeek 实体抽取兜底
- `resolveEventForAction` 中的自然语言匹配步骤

Go 侧 SHALL 直接使用 Python 返回的 `quantity`、`event_id`，根据 `event_id` 查询 Redis 缓存获取 `event_type` 和 `event_unit`。

#### Scenario: 已有事件直接使用 Python 返回数据

- **WHEN** Python 返回：
  ```json
  {
    "event_id": "123",
    "event_name": "喂奶",
    "quantity": 120,
    "is_new_event": false
  }
  ```
- **THEN** Go 侧不调用 `extractNumberFromText`
- **AND** Go 侧根据 event_id 查询 Redis 缓存获取 event_type 和 event_unit
- **AND** Go 侧直接使用 quantity = 120 写入历史记录

#### Scenario: 新事件使用 Python 推断的类型和单位

- **WHEN** Python 返回：
  ```json
  {
    "event_name": "游泳",
    "event_type": "time",
    "event_unit": "分钟",
    "quantity": 30,
    "is_new_event": true
  }
  ```
- **THEN** Go 侧使用 Python 返回的 event_type 和 event_unit 创建事件
- **AND** Go 侧不调用 `callDeepSeekEntityExtract` 进行实体抽取
- **AND** 创建成功后更新 Redis 缓存

### Requirement: Go 侧兼容旧版 Python 返回格式

Go 侧 `AnalyzeIntentResponse` 新增字段 SHALL 使用 `omitempty` 标签，兼容旧版 Python 返回格式。

#### Scenario: 旧版 Python 不返回新字段

- **WHEN** Python 返回：
  ```json
  {
    "target_type": "feeding",
    "action": "one",
    "event_name": "喂奶"
  }
  ```
- **AND** 未返回 event_id、quantity、event_type、event_unit 字段
- **THEN** Go 侧不因缺少字段而报错
- **AND** Go 侧回退到原有逻辑（查询事件列表、提取数量）

## REMOVED Requirements

### Requirement: Go 侧自然语言匹配事件

**Reason**：Python 向量匹配已能高置信度识别事件，Go 侧自然语言匹配逻辑冗余

**Migration**：使用 Python 返回的 `event_id` 直接查询 Redis 缓存

### Requirement: Go 侧 DeepSeek 实体抽取兜底

**Reason**：与"AI 能力全部走 Python 接口"目标冲突

**Migration**：新事件由 Python 推断 event_type 和 event_unit，Go 侧负责入库
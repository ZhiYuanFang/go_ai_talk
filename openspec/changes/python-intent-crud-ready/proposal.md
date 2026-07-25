## Why

当前 Python 意图分析只返回 `event_name` 和 `event_id`，Go 侧仍需：
1. 调用 `extractNumberFromText` 提取数量
2. 通过事件名查找/创建事件获取 `event_type` 和 `event_unit`
3. 保留自然语言匹配逻辑作为兜底

这导致 Go 侧承担了"思考者"角色，违背了"AI 能力全部走 Python 接口"的架构目标。同时，Python 向量匹配已能高置信度识别事件，但无法直接用于 CRUD，因为缺少数量和事件类型信息。

## What Changes

### Python 侧变更

1. **新增返回字段**：
   - `quantity`：从用户输入中提取的数量值（如 "喝了120ml" → 120）
   - `event_type`：事件类型（新事件时返回，值为 number/time/one）
   - `event_unit`：事件单位（新事件时返回，如 ml、次、分钟）

2. **前置数量提取**：
   - 在向量匹配后、LLM 分类前，用正则表达式提取数量
   - 高置信度匹配（>0.95）且能提取数量时，直接返回，跳过 LLM 调用
   - 避免通用场景请求 LLM 导致接口延迟

3. **新事件处理**：
   - 当向量匹配无法找到已有事件时，返回 `is_new_event: true`
   - LLM 根据 Prompt 指导推断 `event_type` 和 `event_unit`

### Go 侧变更

1. **扩展响应结构**：`AnalyzeIntentResponse` 新增 `event_id`、`quantity`、`event_type`、`event_unit` 字段

2. **简化 CRUD 逻辑**：
   - 已有事件：用 `event_id` 查询 Redis 缓存获取 `event_type` 和 `event_unit`
   - 新事件：使用 Python 返回的 `event_type` 和 `event_unit` 入库，并更新 Redis 缓存
   - 直接使用 Python 返回的 `quantity`，删除 `extractNumberFromText` 函数

3. **删除 Go 侧自然语言匹配逻辑**：
   - 删除 `resolveEventForAction` 中的自然语言匹配步骤（`extractEventFromCandidates`、`hasSignificantOverlap`）
   - 删除 DeepSeek 实体抽取兜底（`callDeepSeekEntityExtract`）
   - 保留 `intent.EventName` 匹配和创建新事件逻辑

## Capabilities

### New Capabilities

- `python-intent-crud-ready`：Python 意图分析返回可直接用于 CRUD 的完整数据，Go 侧无需额外解析和推断

### Modified Capabilities

- `voice-internal-text-chat`：修改语音对话能力规格，删除 Go 侧数量提取和自然语言匹配逻辑，依赖 Python 返回的完整 CRUD 数据

## Impact

### Python 侧（D:\work\python_ai_talk）

- `app/feeding/schemas/intent.py`：`IntentResponse` 新增 `quantity`、`event_type`、`event_unit`、`is_new_event` 字段
- `app/feeding/graphs/nodes/prompts/intent_classification.py`：提示词新增数量和事件类型说明
- `app/feeding/graphs/nodes/classify_intent.py`：处理新增字段
- `app/feeding/graphs/nodes/match_event_by_vector.py`：新增前置数量提取逻辑
- `app/feeding/graphs/intent_graph.py`：可能需要调整节点顺序

### Go 侧（D:\work\go_ai_talk）

- `internal/services/voice/python_ai_client.go`：`AnalyzeIntentResponse` 新增字段映射
- `internal/services/voice/voice_chat_understanding.go`：
  - 删除 `extractNumberFromText` 函数
  - 简化 `resolveEventForAction` 函数
  - 删除 `extractEventFromCandidates`、`hasSignificantOverlap` 等自然语言匹配函数
  - 新增根据 `event_id` 查询 Redis 缓存逻辑
- `internal/services/voice/event_tree.go`：删除 `sortForMatch` 等辅助函数
- `internal/services/voice/event_child_pending.go`：调整调用方式
- `internal/services/device/admin.go`：可能需要新增根据 `event_id` 查询 Redis 的方法

### 跨项目协调

- 本变更需要同时修改 Python 和 Go 两个项目
- Python 变更需要先部署，Go 变更才能生效
- 建议先在测试环境验证 Python 新返回格式，再同步部署
# Python Intent CRUD Ready — 任务清单

## 变更概述
本次变更旨在让 Python 侧 LangGraph 返回完整的 CRUD 数据（quantity、event_id、event_type、event_unit、is_new_event），Go 侧据此删除数量提取和自然语言匹配逻辑，统一依赖 Python 接口。

## Python 侧任务（app/feeding/）

### 任务组 1：新增 Schema 字段（app/feeding/schemas/intent.py）
- [x] 1.1 在 `IntentEvent` 中新增 `quantity: Optional[int]` 字段
- [x] 1.2 在 `IntentResponse` 中新增 `quantity: Optional[int]` 字段
- [x] 1.3 在 `IntentResponse` 中新增 `event_type: Optional[str]` 字段
- [x] 1.4 在 `IntentResponse` 中新增 `event_unit: Optional[str]` 字段
- [x] 1.5 在 `IntentResponse` 中新增 `is_new_event: Optional[bool]` 字段

### 任务组 2：新增数量提取工具（app/feeding/utils/quantity_extractor.py）
- [x] 2.1 创建 `quantity_extractor.py` 模块
- [x] 2.2 实现汉字数字到阿拉伯数字的转换
- [x] 2.3 实现正则表达式提取阿拉伯数字
- [x] 2.4 实现 `extract_quantity_from_text` 函数

### 任务组 3：修改向量匹配节点（app/feeding/graphs/nodes/match_event_by_vector.py）
- [x] 3.1 导入 `extract_quantity_from_text`
- [x] 3.2 在向量匹配后调用数量提取
- [x] 3.3 高置信度匹配（>0.95）时，将 `quantity` 加入 `intent_result`
- [x] 3.4 将 `match_source` 放入 `intent_result` 内，供路由节点正确读取
- [x] 3.5 中等置信度匹配（0.90~0.95）时，同样返回 `quantity`

### 任务组 4：修改意图分类节点（app/feeding/graphs/nodes/classify_intent.py）
- [x] 4.1 导入 `extract_quantity_from_text`
- [x] 4.2 处理新事件场景：当事件不在字典中时，标记 `is_new_event=true` 并设置默认 `event_type`/`event_unit`
- [x] 4.3 优先使用向量匹配已提取的 `quantity`，未提取到时本地提取作为 fallback
- [x] 4.4 确保所有新字段都有默认值
- [x] 4.5 多事件场景中也处理 `quantity` 提取

### 任务组 5：修改意图分类提示词（app/feeding/graphs/nodes/prompts/intent_classification.py）
- [x] 5.1 在 prompt 中新增 `quantity` 字段说明
- [x] 5.2 在 prompt 中新增 `event_type`、`event_unit` 字段说明
- [x] 5.3 在 prompt 中新增 `is_new_event` 字段说明
- [x] 5.4 在 prompt 中新增事件类型推断规则

### 任务组 6：修复图路由（app/feeding/graphs/intent_graph.py）
- [x] 6.1 修复 `_route_after_vector_match` 从 `state` 顶层读取 `match_source`（原代码从 `intent_result` 读取导致始终路由到 LLM）

---

## Go 侧任务（hello/）

### 任务组 7：响应结构变更（internal/services/voice/python_ai_client.go）
- [x] 7.1 在 `IntentEvent` 中新增 `Quantity *int` 字段
- [x] 7.2 在 `AnalyzeIntentResponse` 中新增 `EventId string` 字段
- [x] 7.3 在 `AnalyzeIntentResponse` 中新增 `Quantity *int` 字段
- [x] 7.4 在 `AnalyzeIntentResponse` 中新增 `EventType string` 字段
- [x] 7.5 在 `AnalyzeIntentResponse` 中新增 `EventUnit string` 字段
- [x] 7.6 在 `AnalyzeIntentResponse` 中新增 `IsNewEvent bool` 字段

### 任务组 8：简化 resolveEventForAction（internal/services/voice/event_child_pending.go）
- [x] 8.1 删除步骤 1：自然语言匹配（`resolveEventLeaf` 调用）
- [x] 8.2 保留步骤 2：`intent.EventName` 匹配
- [x] 8.3 删除步骤 3：DeepSeek 实体抽取兜底（`callDeepSeekEntityExtract` 调用）
- [x] 8.4 保留步骤 4：新事件创建（`InsertOrGetEventByNeedle`）
- [x] 8.5 修改新事件创建逻辑，使用 Python 返回的 `event_type` 和 `event_unit`

### 任务组 9：删除冗余函数与清理（internal/services/voice/）
- [x] 9.1 删除 `extractNumberFromText` 函数（已删除）
- [x] 9.2 保留 `extractEventFromCandidates`（服务于 `continuePendingChildEvent`）
- [x] 9.3 保留 `hasSignificantOverlap`（服务于 `continuePendingChildEvent`，已恢复）
- [x] 9.4 保留 `sortForMatch`（服务于 `continuePendingChildEvent`）
- [x] 9.5 删除 `callDeepSeekEntityExtract` 函数
- [x] 9.6 删除 `callDeepSeekRaw` 函数
- [x] 9.7 删除 `handleActionRecord` 函数
- [x] 9.8 删除 `resolveEventFromUnifiedIntent` 函数
- [x] 9.9 删除 `chatWithResult` 中的预设动作快速路径循环
- [x] 9.10 删除未使用的 `regexp`、`sort`、`strconv` import

### 任务组 10：编译与全局验证
- [x] 10.1 运行 `go build ./...` 确保编译通过
- [x] 10.2 运行 `go vet ./internal/...` 确保无警告
- [x] 10.3 全局搜索确认无 `extractNumberFromText`、`callDeepSeekEntityExtract`、`callDeepSeekRaw`、`handleActionRecord`、`resolveEventFromUnifiedIntent` 残留引用

---

## 变更文件清单

### Python 侧
| 文件 | 变更类型 |
|------|---------|
| `app/feeding/schemas/intent.py` | 修改（新增字段） |
| `app/feeding/utils/quantity_extractor.py` | 新增 |
| `app/feeding/graphs/nodes/match_event_by_vector.py` | 修改（数量提取、路由字段修复） |
| `app/feeding/graphs/nodes/classify_intent.py` | 修改（新事件处理、数量 fallback） |
| `app/feeding/graphs/nodes/prompts/intent_classification.py` | 修改（prompt 新增字段说明） |
| `app/feeding/graphs/intent_graph.py` | 修改（路由 bug 修复） |

### Go 侧
| 文件 | 变更类型 |
|------|---------|
| `internal/services/voice/python_ai_client.go` | 修改（新增字段） |
| `internal/services/voice/voice_chat_understanding.go` | 修改（删除函数、清理 import） |
| `internal/services/voice/event_child_pending.go` | 修改（简化 resolveEventForAction） |
| `internal/services/voice/voice_chat_deepseek.go` | 修改（删除函数） |
| `internal/services/voice/event_tree.go` | 修改（恢复 hasSignificantOverlap） |

---

## 备注
- 所有 Go 侧编译错误已修复，`go build` 和 `go vet` 均通过
- Python 侧 `py_compile` 语法检查通过
- `extractEventFromCandidates`、`hasSignificantOverlap`、`sortForMatch`、`resolveEventLeaf` 保留，因其服务于 `continuePendingChildEvent` 子事件消歧流程，与 `resolveEventForAction` 是不同功能路径
- 预设动作快速路径已完全移除，所有请求统一走 Python 接口
- 向量匹配节点路由 bug 已修复：高置信度（≥0.95）匹配时正确跳过 LLM

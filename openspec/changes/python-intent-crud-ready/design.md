## Context

当前架构中，Python 意图分析服务负责识别用户意图和事件，但返回信息不完整，导致 Go 侧仍需额外处理：

```
用户输入 → Python 意图分析 → Go 侧处理
              ↓                    ↓
        返回 event_name      查询事件列表
        返回 event_id        提取数量
                            判断事件类型
                            创建/查找事件
```

**当前问题**：
1. Go 侧承担"思考者"角色，违背"AI 能力全部走 Python 接口"的目标
2. `extractNumberFromText` 在 Go 侧实现，重复了 Python 侧已有的文本解析能力
3. 向量匹配置信度已很高（>0.95），但仍需调用 LLM 获取完整信息
4. Go 侧保留了自然语言匹配逻辑作为兜底，与 Python 能力重复

## Goals / Non-Goals

**Goals：**
1. Python 返回可直接用于 CRUD 的完整数据（event_id、quantity、event_type、event_unit）
2. Go 侧删除 `extractNumberFromText` 和自然语言匹配逻辑
3. 已有事件通过 Redis 缓存快速获取 event_type 和 event_unit
4. 新事件由 Python 推断类型和单位，Go 侧负责入库和缓存更新
5. 高置信度匹配时跳过 LLM 调用，降低延迟

**Non-Goals：**
1. 不修改 Python 向量匹配算法本身
2. 不修改 Go 侧 Redis 缓存的整体架构（使用现有 `deviceCache`）
3. 不处理多事件场景（multi-action）的变更（本次暂不涉及）

## Decisions

### Decision 1: 前置数量提取策略

**选择**：在 Python 向量匹配后、LLM 分类前，用正则表达式提取数量

**理由**：
- 避免通用场景请求 LLM 导致接口延迟
- 正则表达式提取数量的准确率很高（当前 Go 侧 `extractNumberFromText` 已验证）
- 用户输入格式相对固定（如"喝了120ml"、"喂了90"）

**备选方案**：
- 方案 A：让 LLM 解析数量 ❌ 增加 LLM 调用开销
- 方案 B：Go 侧保留 `extractNumberFromText` ❌ 重复逻辑，违背目标

**实现细节**：
```python
def extract_quantity_from_text(text: str) -> Optional[int]:
    """从文本中提取数量"""
    # 汉字数字转换
    text = text.replace("一", "1").replace("二", "2")...
    # 正则提取
    match = re.search(r'\d+', text)
    return int(match.group()) if match else None
```

### Decision 2: 已有事件的 event_type/event_unit 获取方式

**选择**：Go 侧根据 event_id 查询 Redis 缓存

**理由**：
- 避免在 Python 侧存储完整事件字典（数据同步问题）
- Go 侧已有 `deviceCache.getEventOptions` 基础设施
- Redis 缓存命中率高，性能优于跨服务查询

**备选方案**：
- 方案 A：Python 返回完整事件信息 ❌ 需要在 Python 侧维护事件字典同步
- 方案 B：Go 侧查询 MySQL ❌ 性能较差

### Decision 3: 新事件的处理方式

**选择**：Python 返回推断的 event_type 和 event_unit，Go 侧负责入库

**理由**：
- LLM 有能力根据上下文推断事件类型（如"游泳" → time 类型，"喝水" → number 类型）
- Go 侧负责数据持久化和缓存更新，符合服务边界划分

**备选方案**：
- 方案 A：Go 侧调用 LLM 推断 ❌ 增加复杂度
- 方案 B：Python 直接创建事件 ❌ 违背服务边界（事件管理属 Go device-service）

### Decision 4: Go 侧删除代码的范围

**选择**：删除以下代码
- `extractNumberFromText` 函数
- `extractEventFromCandidates`、`hasSignificantOverlap`、`sortForMatch` 等自然语言匹配函数
- `callDeepSeekEntityExtract` DeepSeek 实体抽取兜底
- `resolveEventForAction` 中的自然语言匹配步骤

**保留**：
- `resolveEventForAction` 中的 `intent.EventName` 匹配逻辑
- `resolveEventForAction` 中的创建新事件逻辑
- `InsertOrGetEventByNeedle`（新事件创建入口）

**理由**：
- 自然语言匹配已由 Python 向量匹配替代
- DeepSeek 实体抽取兜底与"AI 能力走 Python"目标冲突

## Risks / Trade-offs

### Risk 1: Python 侧数量提取不准确
**影响**：用户输入"一二三"等特殊格式可能无法正确识别
**Mitigation**：
- 保留汉字数字转换逻辑（一→1、二→2...）
- 复杂格式回退到 LLM 解析

### Risk 2: 新事件类型推断错误
**影响**：Python 推断的 event_type 与用户预期不符
**Mitigation**：
- 在提示词中提供常见事件类型推断规则
- 支持后续修改事件类型

### Risk 3: Redis 缓存不一致
**影响**：新事件入库后 Redis 缓存未及时更新
**Mitigation**：
- 使用现有的 `device-event-cache-rebuild-on-mutate` 机制
- 确保新事件创建后触发缓存重建

### Risk 4: 跨项目部署顺序依赖
**影响**：Python 变更未部署时，Go 侧变更会因缺少字段而失败
**Mitigation**：
- Go 侧响应结构新增字段使用 `omitempty`，兼容旧版 Python
- 先部署 Python，验证新返回格式后再部署 Go

## Migration Plan

### Phase 1: Python 变更（先部署）
1. 修改 `IntentResponse` Schema，新增 `quantity`、`event_type`、`event_unit` 字段
2. 修改意图分类提示词，指导 LLM 返回新字段
3. 新增前置数量提取逻辑
4. 部署 Python 服务

### Phase 2: Go 变更（后部署）
1. 修改 `AnalyzeIntentResponse`，新增字段映射（使用 `omitempty`）
2. 简化 `resolveEventForAction`，删除自然语言匹配步骤
3. 删除 `extractNumberFromText` 等函数
4. 部署 Go 服务

### Rollback Plan
- Go 侧保留对旧版 Python 返回格式的兼容（通过 `omitempty`）
- 如需回滚，先回滚 Go，再回滚 Python

## Open Questions

1. **多事件场景（multi-action）是否需要同步修改？**
   - 当前方案暂不涉及，可后续迭代

2. **Redis 缓存查询失败时的降级策略？**
   - 建议：回源 MySQL 查询，与现有 `ListEvents` 逻辑一致

3. **新事件的默认 event_type 推断规则是否需要在提示词中明确？**
   - 建议在提示词中提供示例：喂奶/喝水 → number，睡觉/游泳 → time
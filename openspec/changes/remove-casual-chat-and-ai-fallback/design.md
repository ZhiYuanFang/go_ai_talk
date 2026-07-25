## Context

当前 voice-service 的母婴喂养 AI 能力已全部迁移至 Python 微服务（`python_ai_talk`，基于 LangGraph 编排），但 Go 侧仍保留了两套冗余代码：

1. **闲聊模式业务代码**：母婴/闲聊双模式切换逻辑（模式枚举、模式存储、模式切换命令识别、闲聊流式回复等），但当前产品已统一为母婴喂养场景，闲聊模式不再需要。
2. **AI 能力兜底代码**：所有调用 Python 接口的路径都带有 DeepSeek/LLM 直连兜底，但 Python 侧的 LangGraph 编排（向量匹配 + LLM 分类 + 诊疗图）已完整覆盖所有母婴喂养场景，兜底路径长期不会被触发，徒增维护成本。

Python 侧能力覆盖情况：
- 意图分析（feeding/history/suggest/conversation/exit）：`/v1/analyze/intent` → LangGraph intent_graph
- 成长建议：`/v1/analyze/intent` + `text="成长建议"` → call_clinic_agent
- 历史问答：`/v1/analyze/intent` → history 分支 → fetch_history + generate_response
- 胖宝诊疗流式：`/v1/clinic/stream` → LangGraph clinic_graph
- 对话回复：`/v1/analyze/intent` → conversation 分支 → call_clinic_agent

## Goals / Non-Goals

**Goals:**
- 删除闲聊模式全部业务代码，所有对话统一走母婴喂养模式
- 删除 Go 侧 DeepSeek/LLM 直连兜底代码，AI 推理统一走 Python 微服务
- Python 不可用时返回固定降级提示语（"AI 服务暂时不可用，请稍后再试"）
- 一并清理死代码（`callDeepSeekActionExtract` 等未被调用的函数）
- 简化 `HandleTranscriptForStreaming` 接口签名（移除 `mode`、`needCasualStream`）

**Non-Goals:**
- 不修改 `callDeepSeekEntityExtract`（业务兜底，非 AI 能力兜底）
- 不修改 `callDeepSeekRaw`（被 `callDeepSeekEntityExtract` 依赖）
- 不修改 Python 侧代码
- 不引入新的 AI 能力或业务功能
- 不做接口版本升级（本次为特例，直接简化签名）

## Decisions

### 决策 1：Python 不可用时的降级策略

**方案：** 返回固定提示语 "AI 服务暂时不可用，请稍后再试"

**备选方案对比：**

| 方案 | 优点 | 缺点 |
|------|------|------|
| 固定提示语（选中） | 实现简单，行为一致，用户预期明确 | 无 AI 兜底，Python 挂了=AI 全挂 |
| DeepSeek 直连兜底（当前） | Python 挂了仍有基本 AI 能力 | 维护成本高，两套 AI 链路，行为不一致 |
| 透传 HTTP 错误给前端 | 实现最简单 | 用户体验差，错误码不友好 |

**选择理由：** Python 服务已成为 AI 能力的唯一承载，DeepSeek 兜底与 Python 的能力存在差异（如向量匹配、确认流程等），保留兜底反而可能导致行为不一致。固定提示语降级策略简单清晰，且 Python 服务可用性有保障。

### 决策 2：HandleTranscriptForStreaming 签名简化

**方案：** 直接删除 `mode` 和 `needCasualStream` 返回值，不做接口版本升级

**备选方案对比：**

| 方案 | 优点 | 缺点 |
|------|------|------|
| 直接简化（选中） | 代码干净，无冗余字段 | 违反 AGENTS.md 接口版本不变约定 |
| 保留字段恒值 | 兼容旧调用方 | 字段无意义，增加心智负担 |
| 创建 v2 版本接口 | 符合版本约定 | 工作量大，本次为内部接口无必要 |

**选择理由：** `HandleTranscriptForStreaming` 是 voice-service 内部服务接口，调用方仅为 `voice_ws.go` 控制器，无外部依赖。本次为清理简化，不涉及语义变更，作为特例直接修改。

### 决策 3：保留 callDeepSeekEntityExtract

**方案：** 不删除 `callDeepSeekEntityExtract` 和 `callDeepSeekRaw`

**理由：** `callDeepSeekEntityExtract` 不是 AI 能力兜底，而是业务兜底——当 Python 返回的 `event_name` 在 Go 本地事件树中匹配不到时，用 DeepSeek 做实体抽取作为最后防线。这属于业务逻辑的一部分，与 AI 能力兜底性质不同。`callDeepSeekRaw` 被其依赖，也需保留。

## Risks / Trade-offs

### 风险 1：Python 服务单点依赖

**风险：** 删除兜底后，Python 服务不可用将导致所有喂养 AI 功能直接失败。

**缓解：**
- Python 服务部署应保证高可用（多副本 + 健康检查）
- Go 侧返回友好的固定降级提示语，而非透传技术错误
- 监控 Python 服务可用性，异常时告警

### 风险 2：接口签名简化影响调用方

**风险：** `HandleTranscriptForStreaming` 签名变更可能遗漏调用方。

**缓解：**
- 全局搜索该函数所有调用点，确认仅 `voice_ws.go` 使用
- 修改后编译验证通过
- 运行时测试验证

### 风险 3：删除闲聊模式遗漏边缘场景

**风险：** 某些边缘路径可能仍依赖闲聊模式（如测试代码、配置项等）。

**缓解：**
- 全局搜索 `ChatModeCasual`、`casual`、`闲聊` 等关键词
- 删除后编译验证
- 功能测试覆盖主流对话场景

# confirm-ws-adaptation

## 概述

将 python_ai_talk 服务修复后的 confirm 流程适配到 go_ai_talk 的 WS（WebSocket）模式。go 侧在意图分析响应中接收 `need_confirm`、`confirm_message`、`conversation_id` 三个新字段，并在用户回复后通过 `/v1/analyze/intent/confirm` 接口恢复 Python 侧中断的图执行。

## 背景

python_ai_talk 的 `enhance-feeding-intent-vector-matching` 变更引入了向量匹配 + 用户确认两阶段流程：当向量匹配置信度较低或事件存在歧义时，意图图会中断执行并要求用户确认。修复后的 confirm 流程在 `/v1/analyze/intent` 响应中新增三个字段：

- `need_confirm` (bool)：是否需要用户确认
- `confirm_message` (string)：确认提示话术
- `conversation_id` (string)：会话 ID，用于恢复图执行

并新增 `/v1/analyze/intent/confirm` 接口（请求体 `{"conversation_id": "xxx", "user_feedback": "confirm"|"reject"}`），用于恢复被中断的图执行并返回最终意图分析结果。

go_ai_talk 需要在 WS 模式下适配该流程：上一轮意图分析返回 `need_confirm=true` 时，本轮用户输入需先尝试解析为 confirm/reject，命中后调用 Python 侧 confirm 接口恢复图执行。

## 目标

1. 在 `AnalyzeIntentResponse` 结构体中新增 3 个字段，与 Python 响应对齐
2. 新增 `ConfirmIntent` 方法封装对 Python `/v1/analyze/intent/confirm` 接口的调用
3. 在 `chatWithResult` 中实现 pending 状态管理：检测 pending → 解析用户反馈 → 调用 ConfirmIntent → 复用 `handleUnifiedIntentAction`
4. 在 `callDeepSeekUnifiedIntent` 返回后检查 `NeedConfirm`，命中则保存 pending 状态并返回确认话术
5. 所有代码包含详细中文业务逻辑注释
6. 编译通过，无新增测试文件

## 范围

### 涉及文件

- `internal/services/voice/python_ai_client.go`：扩展响应结构体 + 新增 ConfirmIntent 方法
- `internal/services/voice/voice_chat_understanding.go`：扩展意图结构体 + 字段映射 + pending 检查 + NeedConfirm 分支
- `internal/services/voice/voice_confirm_pending.go`：**新建**，pending 状态管理 + 反馈解析

### 不涉及

- WS 层路由/握手逻辑（复用现有 chatWithResult 入口）
- HTTP 路由新增（confirm 走 Python 内部 HTTP 调用，不暴露 go 侧 HTTP 接口）
- 持久化存储（pending 状态使用内存 map，服务重启后丢失，符合业务预期）
- 测试文件编写

## 设计决策

### 1. pending 状态采用全局变量而非挂载到 VoiceService

confirm 流程与设备模式无关，纯上下文缓存。采用包级单例 `pendingConfirmState` 简化调用路径，避免在 `chatWithResult` 中传递 `s` 实例的额外耦合。与 `pythonAIClientFromCfg()` 全局工厂模式保持一致。

### 2. 60 秒懒加载超时清理

`get` 方法在读取时检查 `time.Since(CreatedAt) > 60s`，超时则返回 nil。无需独立 goroutine 定时清理，与 `pendingChild`/`pendingQuantity` 的内存生命周期模型一致。

### 3. confirm 分支放在 chatWithResult 最前面

pending 状态存在时，无论用户输入什么，都先尝试解析为 confirm/reject：
- 命中 confirm/reject：调用 ConfirmIntent 恢复图执行
- 未命中（用户说了无关内容）：清理 pending，落入常规流程（保证用户能继续操作，不会卡在 pending 状态）

### 4. clear 时机：先清理再调用

调用 ConfirmIntent 前先 `clear(deviceNo)`，避免因 Python 调用失败导致 pending 残留。失败时返回错误提示话术，用户需重新发起意图分析。

### 5. ConfirmIntentResponse 复用类型别名

`type ConfirmIntentResponse = AnalyzeIntentResponse`，因为 Python 侧 confirm 接口返回结构与意图分析一致，复用类型避免重复定义。

### 6. 反馈解析使用关键词列表

`parseConfirmFeedback` 使用肯定词/否定词列表，顺序为：先肯定后否定。肯定词列表包含 "是"/"对的" 等会在 "不是" 之前命中，避免误判。

## 非目标

- 不修改 Python 侧 confirm 接口实现（已在 `fix-confirm-flow-resume` 变更中完成）
- 不增加 go 侧 HTTP 路由（WS 模式下 confirm 完全走 Python HTTP 调用）
- 不持久化 pending 状态（重启丢失符合预期，用户重新发起意图分析即可）
- 不编写测试文件（遵循项目约束）
- 不修改闲聊模式（casual）的 LLM 调用路径（confirm 仅在母婴模式 Python 意图分析后产生）

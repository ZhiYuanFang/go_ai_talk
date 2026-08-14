## ADDED Requirements

### Requirement: MCP 喂养工具改经 chat WS 文模式

mcp-service 的喂养对话工具（现名 `baby_feeding_advisor` 或等价）MUST 通过 `/voice/chat/ws` 完成一轮对话：使用 text 输入与 text 输出模态，发送工具入参文本，读取服务端 `answer`（及可选 `thinking_delta`）作为工具返回。MUST NOT 再调用 `DelegateTextChat` 或 `/voice/internal/api/text/chat`。

#### Scenario: 工具调用成功返回回复

- **WHEN** 小智侧调用 MCP 喂养工具并传入非空 `transcript`
- **THEN** mcp-service MUST 对配置的设备建立（或复用）chat WS，以文模式提交该文本，并在收到业务 `answer` 后将其作为工具结果返回

#### Scenario: 空 transcript 仍本地校验

- **WHEN** 工具收到空或仅空白的 transcript
- **THEN** mcp-service MUST 在本地返回错误，MUST NOT 无意义地占用 WS 意图轮次

### Requirement: MCP 不再依赖 history 文本委派

mcp-service MUST NOT import 或调用 `histsvc.DelegateTextChat` / `DelegateTextChatStream` 作为喂养实现。进程依赖与 runbook MUST 改为描述 WS 连接所需基址/密钥（若有），而非 internal text HTTP。

#### Scenario: 无 DelegateTextChat 调用点

- **WHEN** 检索 mcpbridge / mcp-service 喂养工具实现
- **THEN** MUST NOT 存在对 `DelegateTextChat` 的调用

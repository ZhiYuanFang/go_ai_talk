## ADDED Requirements

### Requirement: 删除对外 history 文本喂养 API

系统 MUST NOT 再提供 `POST /device/history/api/chat` 与 `POST /device/history/api/chat/stream` 作为自然语言喂养入口。本变更 MUST 直接删除（或取消注册）上述路由与 `HistoryCtrl` 对应处理方法，MUST NOT 保留过渡期 200 兼容实现。

#### Scenario: chat 路径不可用

- **WHEN** 客户端请求 `POST /device/history/api/chat` 或 `.../chat/stream`
- **THEN** 网关/history MUST 返回未找到或未注册路由的失败语义（非成功 SSE/JSON 对话）

### Requirement: 删除 voice internal 文本喂养 API

voice-service MUST NOT 再提供 `POST /voice/internal/api/text/chat` 与 `POST /voice/internal/api/text/chat/stream`。history-service MUST 删除 `DelegateTextChat` 与 `DelegateTextChatStream` 及其调用点。若 `POST /voice/text/chat` 仅服务于已删除文本喂养调试路径，MUST 一并删除。

#### Scenario: internal text chat 不可用

- **WHEN** 带内部密钥请求 `/voice/internal/api/text/chat` 或 `.../stream`
- **THEN** voice-service MUST 不再成功执行文本喂养对话

### Requirement: 自然语言喂养仅经 chat WS

自然语言喂养（意图理解与业务落地）对外 MUST 仅经 `/voice/chat/ws`。系统 MUST NOT 再维护第二条 HTTP 文本喂养外壳。

#### Scenario: 文档与 Admin 说明一致

- **WHEN** 运维查看 voice-admin 或相关 runbook 中的喂养 AI 入口列表
- **THEN** 列表 MUST 以 `/voice/chat/ws` 为喂养对话入口，MUST NOT 再将已删的 history/internal text chat 路径列为可用入口

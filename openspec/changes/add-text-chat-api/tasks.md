## 1. Implementation
- [x] 1.1 新增 API 模型：`TextChatReq`（`deviceNo`、`text`）与 `TextChatRes`（`reply`）。
- [x] 1.2 新增控制器与路由：POST `/text/chat`，接入现有鉴权中间件（`Token`）。
- [x] 1.3 复用现有 DeepSeek + 会话缓存逻辑：文字接口与语音接口按同一 `deviceNo` 共享历史上下文。
- [x] 1.4 明确失败语义：DeepSeek 失败时不写入本轮历史；成功后写入 user/assistant 轮次。
- [x] 1.5 文档补充：README 中新增 `/text/chat` 用法与示例，说明与 `/voice/chat` 的跨端共享。

## 2. Validation
- [x] 2.1 单元测试：首次对话、同设备多轮、不同设备隔离、TTL 过期重置。
- [x] 2.2 跨端共享测试：同一 `deviceNo` 先文字再语音（或反之）应携带对方产生的历史上下文（验证 DeepSeek messages 条数变化）。
- [x] 2.3 回归测试：现有 `/voice/chat` 行为不变。

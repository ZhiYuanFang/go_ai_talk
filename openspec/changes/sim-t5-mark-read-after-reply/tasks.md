## 1. ucg-service

- [x] 1.1 `ProcessOutboundChatMessage` 返回投递 `ChatMessage`（WS 调用方忽略返回值）
- [x] 1.2 `ucgInternalChatSend`：send 成功后 `MarkConversationRead(senderWxId, convId, msg.ID)`

## 2. 验收

- [x] 2.1 `go build ./cmd/ucg-service/...` 通过
- [x] 2.2 `openspec validate sim-t5-mark-read-after-reply` 通过

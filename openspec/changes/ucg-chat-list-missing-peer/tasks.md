## 1. 列表 peer 容错

- [x] 1.1 新增 `lookupPeerWxIDOptional`（无对方 member 行时返回 `0, false`，不报错）
- [x] 1.2 `loadConversationDTO` 增加 `tolerateMissingPeer`：`ListConversations` 传 true；有 member 行时保留 `peerWxId`，`ValidateWx` 不存在则跳过 profile 填充
- [x] 1.3 `GetOrCreateDirectConversation` 保持严格 peer 校验（`tolerateMissingPeer=false`）

## 2. 验证

- [ ] 2.1 手工或联调：对方 member 缺失时会话列表 200 且单项 peer 展示为空
- [ ] 2.2 手工或联调：对方 wx 已注销时 `peerWxId` 保留历史 id、nickname/avatar 为空；WS 发消息对无对方会话仍失败

## Context

- `ListConversations` 对每条 membership 调用 `loadConversationDTO`，后者调用 `peerWxID`；成员行缺失时返回 `CodeNotFound`「会话成员不存在」，循环 `return nil, err` 导致整页失败。
- `GetPublicProfile` 失败时已静默省略昵称/头像，不是问题点。
- 账号注销（`device/app/api/user/deactivate`）当前仅删 `wx` 行，不联动 UCG；但运维/未来可能删对方 `ucg_conversation_member`，或 profile/wx 不可用导致不应展示旧资料。
- 用户确认：**列表项 `peerWxId` 保留历史 id**（来自 member 行）；展示字段为空即可。

## Goals / Non-Goals

**Goals:**

- 列表接口对「对方不可用」会话容错，单条降级不影响同页其他会话。
- 有 member 行时保留 `peerWxId`；无 member 行时 `peerWxId=0`。
- 对方 wx 不存在（`ValidateWx.exists=false`）时不填充 nickname/avatar，避免展示已注销用户资料。

**Non-Goals:**

- 注销时联动清理 conversation_member（不在本变更）。
- 修改 WS 发消息、审核、创建会话等路径的 `peerWxID` 硬失败语义。
- 新增 `peerDeleted` 等 API 字段（客户端用空 peer 字段推断）。

## Decisions

### 1. 列表专用 peer 解析

- **决定**：新增 `lookupPeerWxIDOptional(ctx, convID, selfWxID) (peerID uint64, ok bool)`，查不到对方 member 行时 `(0, false)`，不返回 error。
- **决定**：`loadConversationDTO` 增加参数 `tolerateMissingPeer bool`；`ListConversations` 传 `true`，`GetOrCreateDirectConversation` 传 `false`（仍走严格 `peerWxID` 或同等校验）。

### 2. peer 展示填充

```
peerID, ok := lookupPeerWxIDOptional(...)
dto.PeerWxId = peerID   // ok 时保留历史 wx_id；!ok 时为 0

if ok && peerID > 0 {
    if exists, _, vErr := Device().ValidateWx(ctx, int64(peerID)); vErr == nil && exists {
        if prof, pErr := GetPublicProfile(ctx, peerID); pErr == nil && prof != nil {
            // 填充 peerNickname / avatar*
        }
    }
}
// 否则 peer 展示字段保持空
```

- **理由**：member 行在但 wx 已删时，保留 `peerWxId` 供点进会话看历史，但不展示 stale profile。

### 3. 严格路径不变

- `peerWxID` 保留，供 `ProcessOutboundChatMessage`、审核 reject 等使用；行为不变。

## Risks / Trade-offs

- [列表多一次 ValidateWx/peer] → 仅对当前页会话（默认 ≤50）调用；可接受。
- [peerWxId=0 与注销共存] → 仅 member 行完全缺失时发生；客户端占位即可。

## Migration Plan

- 单服务 ucg-service 部署即可；无数据迁移。
- 回滚：还原 `loadConversationDTO` 严格行为。

## Open Questions

- 无。

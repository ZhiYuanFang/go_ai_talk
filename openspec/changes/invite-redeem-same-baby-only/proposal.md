## Why

产品决定邀请码互兑闸从「不可使用自己的邀请码（同 `wx_id`）」改为「不可使用同一宝宝下账号的邀请码（同 `device_no`）」。现网仅保留自用闸，同宝宝闸曾被删除，导致配偶等同机账号可互兑；同时规格仍要求自用+同宝宝双闸，与产品选项 A（只留同宝宝）不一致，需对齐实现与规格。

## What Changes

- **BREAKING（兑码业务语义）**：移除 `owner_wx_id == redeemer_wx_id` 自用拒绝；允许同一人在**不同宝宝**上兑换自己的邀请码。
- 恢复并作为唯一家庭互兑闸：**同宝宝禁兑**——码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空则拒绝；文案对齐「不可使用同一宝宝下其他账号的邀请码」（自兑同机也命中同宝宝，不再单独报「不可使用自己的邀请码」）。
- 主人设备号经 device 契约查询；查询失败 fail-closed；主人未绑机（空 `device_no`）不因同宝宝规则拒绝。
- 更新 Admin 只读规则文案与 OpenSpec：去掉「不可自用」MUST，保留/强化同宝宝。
- 保留人×码×功能去重、InviteOncePerUser 等其它闸；**不恢复**一家锁定 / InviteOncePerDevice。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `invite-code-identity`：兑码去重要求去掉自用，仅保留同宝宝 + 既有人×码×功能等。
- `feature-admin-rule-readonly`：共用规则文案去掉「不可使用自己的邀请码」，保留同宝宝说明；同宝宝禁兑要求保持生效。
- `feature-unlock-fulfillment`：邀请码兑换规则去掉「不可自用」；与同宝宝闸对齐（不再写一家绑定为现行 MUST）。
- `growth-invite-per-user-only`：去掉「MUST 保留自用禁兑」，保留同宝宝与人维一次等。

## Impact

- **cash-service**：`internal/services/cash/feature_invite.go`（`RedeemInviteCode`）、`feature_admin_rules.go` 文案
- **clients**：`internal/clients/device.FetchDeviceNoByWxID`（恢复调用）
- **规格**：上述四个 capability 增量
- **不涉及**：新 App 路由、Redis 新读缓存、背景 ticker；gateway usage 策略不变（仍走既有 redeem）

## Context

`RedeemInviteCode`（`feature_invite.go`）当前仅拒绝 `owner_wx_id == redeemer_wx_id`。同宝宝闸（主人当前 `device_no` vs 兑换者 header `device_no`）已从实现删除，但 v3.0.3 / Admin 文案仍描述双闸。产品选项 **A**：去掉自用，只留同宝宝。

兑换者 `deviceNo` 已由网关头强制非空；主人 `device_no` 须经 `deviceclient.FetchDeviceNoByWxID`（`DEVICE_SERVICE_URL`），cash MUST NOT 直查 device 库。

## Goals / Non-Goals

**Goals:**

- 兑码路径：删除自用比较与「不可使用自己的邀请码」错误。
- 恢复同宝宝拒绝：双方 `device_no` 非空且相等 →「不可使用同一宝宝下其他账号的邀请码」（或同等语义）。
- device 查询失败 fail-closed；主人未绑机不因同宝宝拒绝（此时自兑自己的码若主人未绑机将**允许**，属选项 A 刻意结果）。
- Admin 共用规则文案与四份规格去掉「不可自用」MUST。

**Non-Goals:**

- 不恢复一家锁定、`feature_invite_redeemer_bind`、InviteOncePerDevice。
- 不改人×码×功能去重、InviteOncePerUser、获客原力通知。
- 不新增 App 路由 / Redis 读缓存 / 背景循环。

## Decisions

### D1：用同宝宝完全替换自用（选项 A）

- **选择**：不再比较 `wx_id`；同机（含自己兑自己且同绑）一律走同宝宝文案。
- **备选 B（自用+同宝宝）**：否决（用户明确 A）。
- **后果**：主人未绑机或与兑换者不同 `device_no` 时，同一 `wx_id` 可兑自己的码。

### D2：同宝宝判定口径

- **选择**：`ownerDeviceNo == redeemerDeviceNo` 且二者 `TrimSpace` 后均非空；主人设备取**当前**绑定（非历史流水）。
- **备选**：按「一家账号列表」包含主人 — 等价于同 `device_no` 时更重，否决。

### D3：校验时机

- **选择**：TX 外 peek 阶段在码有效后查询主人 device 并做同宝宝判断（与历史实现一致，失败早退）；TX 内锁码后再做一次同宝宝（防绑机变更竞态可选；至少 TX 内保留一次权威判断）。实现最小集：TX 内锁码后查主人 device + 同宝宝（若 peek 已查可复用结果但绑机可能变，**建议 TX 内再查一次**或接受 peek 一次以减 RPC——优先 **TX 外一次 + 与锁码后自用位置对称的一次**；为最少 RPC：**仅在锁码成功后查一次**即可）。
- **定稿**：码有效性确认后、人×码×功能去重前，**调用一次** `FetchDeviceNoByWxID(owner)`；失败拒绝；通过后再进其余闸。TX 内外是否双检：与现自用双检对称可在 peek + TX 各一次；若为降 RPC，允许仅 TX 内一次（peek 不再做家庭闸）。**推荐 peek+TX 双检对齐旧自用结构，device 调用两次可接受。**

### D4：文案

- 拒绝：`不可使用同一宝宝下其他账号的邀请码`
- Admin：`共用：不可使用同一宝宝（同 deviceNo）下其他账号的邀请码。`（去掉「不可使用自己的邀请码；」）

## Risks / Trade-offs

- **[自兑跨宝宝被允许]** → 选项 A 预期；运营需知。
- **[device 抖动导致误拒]** → fail-closed；可重试。
- **[主人刚换绑]** → 以查询时刻当前绑定为准。
- **[规格与旧「一家绑定」段落残留]** → 本变更 MODIFIED `feature-unlock-fulfillment` 邀请兑换条款，去掉自用与过时一家绑定现行 MUST（与 invite-code-identity「MUST NOT 一家锁定」对齐）。

## Migration Plan

1. 发布 cash-service（依赖 device-service 内部 device-no-by-wx-id 可用）。
2. 抽检：同机两账号互兑拒；不同机好友互兑过；同人异机兑自己的码过；device 挂掉时兑失败。
3. 回滚：恢复自用比较、去掉同宝宝查询即可。

## Open Questions

- （无阻塞）产品已选 A。

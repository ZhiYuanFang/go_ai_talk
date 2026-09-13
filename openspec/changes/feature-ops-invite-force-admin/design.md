## Context

商业功能开通已落地三功能：`prediction_unlock`（机·条数）、`care_alert_smart_remind`（机·权益）、`growth_trajectory_predict`（人·权益）。邀请共用规则含自用禁兑、同宝宝禁兑（`invite-block-same-device`）。成长轨迹曾错误叠加 `InviteOncePerDevice`。获客原力 cash→ucg `force/acquire` 在无 `ucg_user_force` 行时 `LockUpdate().Scan` 触发 `sql.ErrNoRows`，兑码成功但加分失败。Admin 仅能编辑定义/SKU，无规则说明与开通快照。

约束：跨服务禁直查；Redis 经 cachekit；不新增背景 ticker；不新增 `*_test.go`；Admin 接口非 App usage；接口结构变更对已有 App v1 行为以兼容为主（成长轨迹邀请放宽为行为修正，非 BREAKING 字段）。

## Goals / Non-Goals

**Goals:**

- 成长轨迹邀请与「对人开通」一致：同机可兑不同码，每人仅一次。
- 修复原力首次加分；补齐 profile/me 原力展示。
- Admin：只读规则 + 方案 A 当前开通快照详情页。
- 明确保留同宝宝禁兑。

**Non-Goals:**

- 开通事件账本（方案 C）或邀请+订单拼装准历史（方案 B）。
- VIP 覆盖写入权益表或详情页完整 VIP 合成。
- 自动后台补偿任务；Flutter 强制改版（除非文案需对齐）。
- 修改预测/值得留意的 InviteOncePerDevice 语义。

## Decisions

### D1：成长轨迹仅人闸

- **选择**：`InviteOncePerDevice` 仅保留 `care_alert_smart_remind`；成长轨迹只保留 `InviteOncePerUser` + 人×码×功能。
- **存量**：兑码校验与写入均跳过成长轨迹的 `feature_invite_device_grant`；旧行可保留不删（忽略即可）。
- **替代**：迁移删除旧行——非必须，冷数据可忽略。

### D2：同宝宝禁兑不变

- 继续：主人当前 `device_no`（device 契约）与兑换者请求头 `deviceNo` 均非空且相等则拒；主人空绑机不因本规则拒；查询失败 fail-closed。
- 文案维持：「不可使用同一宝宝下其他账号的邀请码」。

### D3：原力空行写入

- **选择**：`AddForceDelta` 用 `One()` + `IsEmpty()`（或忽略 `ErrNoRows`），空则 INSERT，有则加锁更新；与 device `wx.go`「空集不用 Scan」一致。
- **可选**：`INSERT ... ON DUPLICATE KEY UPDATE force_value = force_value + ?` 单语句。
- **补偿**：一期提供运维文档/一次性 SQL 或 Admin 只读指引：按 `feature_invite_redemption` 与 ledger 缺口补 `invite_acquisition` +100；**禁止**新增 ticker。默认不自动跑全量补偿（防重复），以「ledger 无该 code ref 则补」幂等为准若实现脚本。

### D4：profile/me enrich

- `mergeProfileForAuthor`（或 GetOrCreateMyProfile 返回前）调用既有 `enrichProfileForceValues`。

### D5：Admin 只读规则

- 服务端按 `featureId` 返回 `ruleSummary`（或静态映射常量），Admin HTML 只展示不可编辑。
- 三功能文案含：主体、邀请规则、付费、同宝宝禁兑共用句；成长轨迹用「同机可兑不同码、每人一次」。

### D6：Admin 开通快照（方案 A）

- 新 Admin API：按 `featureId` 分页列出当前行：
  - `device` 功能：`feature_entitlement`（deviceNo、unlockMethod、expiresAt、active、remainingSeconds 或文案）。
  - `user` 功能：`feature_user_entitlement`（wxId、同上；昵称尽力经 ucg batch，失败可空）。
  - `prediction_unlock`：`feature_allowed_count`（deviceNo、permanentDelta；状态「条数有效」、无到期概念或 N/A）。
- 已过期行仍可列出并标 `active=false`。
- 页内注明：不含 VIP 旁路；非完整开通事件历史。
- UI：列表「详情」进入详情区/子页；上部规则只读，下部快照表。

### D7：usage / 路由

- 仅 Admin 路径；不问 App usage；gateway Admin 静态与 cash Admin 反代按既有模式。

## Risks / Trade-offs

- [成长轨迹旧 device_grant 若误校验] → 代码路径显式排除 featureId。
- [补偿重复加分] → 以 ledger `ref=code` + reason 幂等；或人工审后执行。
- [方案 A 看不到续期前渠道] → 接受；文档写明；后续可开 B/C。
- [预测无「人」列] → 详情只展示 device + 条数；邀请人可另期做。

## Migration Plan

1. 先发 ucg 原力修复（止血加分）+ profile enrich。
2. 再发 cash 成长轨迹闸 + Admin API/页。
3. 按需执行历史原力补偿（运维窗口）。
4. 回滚：常量与函数可回退；Admin 只读接口可下线页入口。

## Open Questions

- 历史原力补偿是否纳入本变更实现脚本，还是仅文档？（默认：实现幂等补偿入口或 SQL 文档二选一，tasks 标可选。）

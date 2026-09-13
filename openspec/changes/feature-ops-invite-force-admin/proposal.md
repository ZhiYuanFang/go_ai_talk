## Why

成长轨迹已按账号开通，但仍套用「同机邀请仅一次」，导致同宝宝其他家长无法各自邀请开通；获客原力因 `ucg_user_force` 空行 `Scan` 返回 `sql.ErrNoRows` 导致首次加分失败（兑码成功但积分/流水为空）；运维在功能管理后台看不到各功能开通规则与当前开通主体状态，排查成本高。

## What Changes

- **成长轨迹邀请闸**：`growth_trajectory_predict` 移出 `InviteOncePerDevice`；保留 `InviteOncePerUser` 与人×码×功能去重；同宝宝可兑多个不同好友码；存量 `feature_invite_device_grant` 对成长轨迹停止校验/停写（或忽略该 feature 旧行）。
- **共用邀请规则重申**：兑码继续拒绝码主人与兑换者当前 `device_no` 相同（同宝宝）；文案保持「不可使用同一宝宝下其他账号的邀请码」；device 查询失败 fail-closed。
- **原力首次加分修复**：`AddForceDelta` 空集不得因 `Scan`/`ErrNoRows` 失败；须能 INSERT 首行并写 ledger；可选按 cash 邀请兑换记录补偿历史未加分（设计定夺，默认一次性脚本或 Admin/运维 SQL 指引，**不**新增背景 ticker）。
- **profile 原力展示（次要）**：`profile/me` 路径 enrich `forceValue`，与公开主页/ledger 一致。
- **Admin 功能只读规则**：开通功能管理对每个 `featureId` 展示固定只读开通规则说明（主体、邀请、付费、同宝宝禁兑），不可编辑。
- **Admin 开通详情（方案 A·当前态）**：每个功能可进入详情页，列出当前权益/预测条数主体的开通方式、是否有效、剩余时效（或永久）；**不**新建事件账本；不含 VIP 旁路覆盖说明为主（页内注明本页不含 VIP）。

## Capabilities

### New Capabilities

- `growth-invite-per-user-only`：成长轨迹邀请仅按人一次，允许同机多码；修正与 `activation_subject=user` 一致。
- `ucg-force-first-grant`：原力增量空行可首次写入；获客/辩论首笔不再因 ErrNoRows 失败；可选历史补偿策略。
- `feature-admin-rule-readonly`：功能定义管理展示只读开通规则文案。
- `feature-admin-activation-snapshot`：Admin 按功能查看当前开通主体快照（方式、状态、剩余时效）。

### Modified Capabilities

- （无主库 `openspec/specs/` 基线 delta 强制项；行为以本变更新 capability 与既有 `invite-block-same-device` / `growth-trajectory-user-entitlement` 变更规格对齐修订为准。）

## Impact

- **cash-service**：`InviteOncePerDevice`、`RedeemInviteCode`、Admin API/静态页、可选补偿查询。
- **ucg-service**：`force_store.AddForceDelta`、`profile/me` enrich。
- **gateway-app**：Admin 静态资源与 Admin 反代（若新 Admin 路径）；无 App Bearer 白名单变更预期。
- **库**：不强制新表（方案 A）；补偿若执行则写 `ucg_user_force` / `ucg_force_ledger`。
- **Flutter**：非必须；原力展示随服务端 enrich 自然修复；兑码同宝宝提示沿用现网文案。
- **非目标**：事件级开通账本（方案 C）；Admin 拼装邀请+订单准历史（方案 B）；新建 `*_test.go`；背景循环补偿任务。

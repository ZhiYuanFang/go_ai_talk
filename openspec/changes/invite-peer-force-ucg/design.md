## Context

商业功能邀请码与预测开通已在 cash 落地（Admin 造码、人×功能一次、一家锁定、预测 T1 全开）；原力在 `wx.force_value` 由 ucg 经 device 内部 API 加减。产品未发布。本变更将邀请码改为用户互推身份、预测三通道按次 +1、原力收回 ucg，并由 device 提供非叶子事件数供 catalog 展示天花板。

约束：跨服务禁止直查他域表；Redis 经 cachekit；不新增背景 ticker；App 新接口须问 usage；不新增测试；接口结构变更走兼容策略（未发布可直接改现有 v1 行为，但不得破坏无关 VIP 契约）。

## Goals / Non-Goals

**Goals:**

- 邀请码身份与功能开通解耦：码常在；能否兑由 `feature_def.unlock_methods` 决定。
- 预测：支付/邀请/广告每次永久 +1；VIP 仅时效覆盖使用权。
- 获客 ×100 → ucg 原力 + ledger；辩论 +1 + ledger。
- catalog 带 `totalActivatableCount`（device 非叶子计数 → cash 聚合）。
- 删除邀请码 Admin 整页及造码入口。

**Non-Goals:**

- 不回填历史投票/获客流水；不做 wx 表存码。
- 不改 UCG 入场喂养门槛；不改 VIP 续期/支付渠道本身（仅澄清与单事件叠加语义）。
- 不在服务端用档位阈值做业务裁决；不返回「距下一档」。
- 不引入新的 Redis 读缓存（除非沿用既有 catalog 失效模式且 design 已说明）。

## Decisions

### D1：邀请码仍在 cash（方案 B）

- **选择**：`feature_invite_code` 一 `owner_wx_id` 一码；注册成功或首次 `GET mine` 时 Ensure。
- **替代**：码写入 wx——拒绝（越界且兑码/获客账本仍在 cash）。
- 废弃字段语义：`expires_at`/`grant_duration_days`/码级功能子表/ `max_redemptions`（或恒 0 不限制）/ `feature_invite_redeemer_bind` 不再参与校验；可删表或停用读写（未发布优先删除无用路径）。

### D2：兑码键与互刷

- 去重：`(redeemer_wx_id, code, feature_id)` 唯一。
- 禁止自用；允许多好友码；允许 A↔B 互兑同一 feature 各一次。
- 预测：`GrantEntitlementOrCount(..., allowed_count_delta, 1)`；**禁止** `GrantPredictionFullAccess`。
- 非预测 + `invite_code`：`entitlement` 一次。

### D3：支付/广告与邀请同为 +1

- 预测 SKU / 广告完成：`grantQuantity=1` 的 `allowed_count_delta`。
- Admin 种子与校验：预测支付商品 grantQuantity MUST 为 1。

### D4：原力迁 ucg

- 表：`ucg_user_force(wx_id PK, force_value, updated_at)`；`ucg_force_ledger(id, wx_id, reason, delta, ref, created_at)`。
- reason：`debate_self_vote`（+1）、`invite_acquisition`（+100）。
- 投票成功写本域；cash Redeem 成功后调 ucg **内部**加分 API（失败记日志并定义是否回滚兑码——推荐：兑码事务提交后异步/同步加分，加分失败可重试或补偿，**不以原力失败回滚已发放的预测条数**，避免用户已得权益被撤销；ledger 以成功写入为准）。
- device：删除 force 列与 increment；BatchWx/Validate 不再返回 forceValue。
- profile enrich：读 ucg 本域；可继续下发 `forceTier` 作展示兜底，客户端以 `forceValue` 本地算档为准。

### D5：totalActivatableCount（方案 C）

- **device**（事件字典权威）：提供内部接口，统计事件表中 **非叶子**节点数（存在至少一条 `parent_id` 指向该事件的子节点）。
- **cash** 合成 catalog 时调用该契约，写入预测项 `totalActivatableCount`；调用失败时字段为 0 或省略并打可观测日志（客户端不得误显示「已全部激活」——建议失败时不返回正数，客户端仅当 total>0 且 activated≥total 才显示全部激活）。
- history 不另维护第二份计数；若读模型缓存事件字典在 history，仍以 device 字典为计数源。

### D6：VIP 解耦

- 写：VIP 履约只动 `vip` 权益表；永不改 `feature_allowed_count`。
- 读（App 使用态）：功能可用 = 永久 entitlement/条数 **OR** `isVip`；catalog 对预测仍返回**永久**合成 `allowedCount`（defaultFree+delta），**不**因 VIP 改成 -1。
- 废除「邀请导致 catalog allowedCount=-1」后，-1 仅保留若仍有其它临时全开来源；本变更目标为预测无 T1。

### D7：Admin 邀请页

- 删除 `cash-invite-code-admin.html`、Hub 模块、静态路由、auth exempt、Admin 造码/列表 API（若仅服务该页）。
- 兑换明细若运维仍需，可后续另开只读工具；本变更不保留造码 UI。

### D8：App API 草案（路径可在实现微调，须 g.Meta）

- `GET /cash/app/api/invite/mine` → `{ code, redeemedCount }`（Ensure）
- `GET /cash/app/api/invite/invitees` → `[{ wxId, nickname, redeemedAt }]`（昵称经 ucg/device 展示契约）
- `POST /cash/app/api/feature/invite-codes/redeem`：规则按 D2（保留路径，改语义）
- `GET /ucg/app/api/force/summary` → `{ forceValue }`（可选，或并入 profile）
- `GET /ucg/app/api/force/ledger` → 分页流水
- `POST /ucg/internal/api/force/acquire`（cash 调用，+100）
- device：`GET|POST /device/internal/api/event/non-leaf-count` → `{ count }`
- 注册：device 创建 wx 后调 cash Ensure（或 cash 在 mine 懒创建）

### D9：usage

- 新增对外 App 接口实现前 **MUST** 询问负责人是否计入 usage；未答复前不改 `maintenance_skip.go`。

## Risks / Trade-offs

- [获客加分与兑码不一致] → 兑码成功优先；加分失败可重试/补偿任务（若引入补偿 MUST 先有 OpenSpec 批准背景任务；一期可同步调用 + 日志告警人工补）。
- [非叶子数与预测展示行不一致] → 锁槽位与「可激活总数」MUST 对齐同一非叶子集合；客户端预测锁下标与 total 使用同一语义。
- [互刷原力] → 产品接受 A↔B 互兑；不设 max_redemptions。
- [删除 Admin 造码] → 仅依赖 Ensure；运维停用可用 DB/status 字段或后续只读工具。

## Migration Plan

1. 部署 device（非叶子计数 + 去 force）→ ucg（新表 + API）→ cash（Redeem/Ensure/catalog）→ gateway。
2. 无需数据迁移；可丢弃未发布的 invite 测试码。
3. 回滚：恢复旧二进制（未发布环境可接受短暂不一致）。

## Open Questions

- 无（探索已锁定：方案 B/C、×100、广告 +1、原力在 ucg、删 Admin 页、不回填流水）。

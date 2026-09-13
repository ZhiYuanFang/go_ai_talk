# 获客原力历史补偿（可选、人工）

## 背景

`AddForceDelta` 曾在 `ucg_user_force` 无行时因空集 `Scan` 返回 `sql.ErrNoRows`，导致兑码成功但获客 +100 未写入。修复后仅保证**新兑码**加分；历史缺口需幂等补账。

## 幂等原则

- ledger 已存在 `reason='invite_acquisition'` 且 `ref=<邀请码>` 的 owner 行 → **跳过**。
- 否则按 cash `feature_invite_redemption`：每个 `(owner_wx_id, code)` 成功兑换组补一次 +100（同一 code 多次兑不同功能仍只应按「获客一次」还是「每次兑一次」？）

现网语义：每次成功兑码（任意功能）都会 `NotifyUcgInviteAcquisition(owner, code)`，**同一 code 多次兑不同功能会多次 +100**。补偿应对齐：**每条 redemption 一行对应一次 +100**，但若历史上同一 code 多次失败且 ledger 无记录，可按 redemption.id 或 `(owner, code, redeemed_at)` 去重；更简单做法：

1. 列出 cash 中每个 owner 的 `redeemed_count` 或 redemption 条数 N。
2. 统计 ucg ledger 中该 owner 的 `invite_acquisition` 条数 M。
3. 若 N>M，差额 `(N-M)*100` 补入余额，并插入 `(N-M)` 条 ledger（ref 可用 `backfill:<redemption_id>` 避免与真实 code 冲突）。

**禁止**上线背景 ticker 自动扫表；本文件仅供运维窗口人工执行。

## 示例核对 SQL

```sql
-- cash：某主人兑换次数
SELECT owner_wx_id, COUNT(*) AS n
FROM feature_invite_redemption
GROUP BY owner_wx_id;

-- ucg：获客流水次数
SELECT wx_id, COUNT(*) AS m
FROM ucg_force_ledger
WHERE reason = 'invite_acquisition'
GROUP BY wx_id;
```

补账须在事务内同时更新 `ucg_user_force` 与插入 `ucg_force_ledger`，ref 建议 `backfill:<id>`。

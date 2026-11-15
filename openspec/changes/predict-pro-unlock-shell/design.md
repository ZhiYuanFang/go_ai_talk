## Context

工作区已实现 `predict-capability-unlock`：双档（basic/pro）、imminent 按 basic∨pro∨VIP 门禁、LA upsert 剥 imminent、Pro 喂养资格、广告清理、catalog `status`。产品改口后：推送必须全员开放；本轮只要 Pro 开通壳（后续提案再挂能力）。

## Goals / Non-Goals

**Goals:**

- 删除基础版商业实体与代码路径。
- 拆除 voice/history/clients 上的预测能力 access 门禁，推送回到 3.1.9。
- 保留 Pro：对人、支付/邀请/7 日点试用、`predict_pro_entry` 前置、eligibility（计入 usage）。
- 保留广告清理与 catalog `status`。

**Non-Goals:**

- 实现任何 Pro 业务能力（分析、增强推送、UI 特权等）。
- 用 VIP/Pro 再门禁 imminent。
- 回退广告清理或 catalog `status`。
- 删除已产生的 Pro 权益历史行（可保留表数据；仅停 basic 种子）。

## Decisions

### D1: 新 change 收窄，而非改写旧提案归档语义

- 本 change `predict-pro-unlock-shell` 为实施真相；对已写代码做**回退 + 收敛**。
- 旧 change 目录可随后 archive/废弃说明，避免双源规格。

### D2: 推送完全不读开通

- 删除 `GetPredictCapabilityAccess`、internal API、`RemotePredictCapabilityAccess`。
- `predict_imminent` 恢复按 wx 直接推；`RefreshBabyLiveActivity` 去掉 strip/access，空投影仍 end。
- **备选（否决）**：仅 VIP 可推——与「开放预测推送」冲突。

### D3: Pro 壳保留方式

- 保留 `predict_capability_pro` Ensure 种子与 SKU、`StartFeatureTrial`（仅 Pro）、建单/邀请 Pro 喂养校验、eligibility API、Hub 规则文案。
- `IsPredictCapabilityFeature` 仅 Pro；删除 basic 常量引用与 schema INSERT。
- Ensure 对 basic：可不 DELETE 历史行；停止 INSERT/强制 UPDATE；Admin 平铺若库中仍有 basic 行可 status=0 或忽略——**推荐 Ensure 将 basic `status=0` 且不再展示依赖**（若行存在）。

### D4: 喂养资格与广告 / catalog

- `predict_pro_entry` 与 eligibility **不动**（仍 Pro 开通前置）。
- 广告拒绝路径、Hub 无 ad 勾选、catalog 无 adDurationDays + 含 status **不动**。

## Risks / Trade-offs

- [已购 basic 用户] → 权益行残留但无业务读取 → 可接受；Hub 可下架 basic。
- [已部署门禁版本] → 未开通用户收不到推送 → 必须重部署 voice+history 回退门禁。
- [Pro 已开通无体感] → 产品预期；catalog unlocked 仍 true，能力后续提案再挂。

## Migration Plan

1. 部署 cash（停 basic 种子/可 status=0；Pro 壳保留）。
2. 部署 voice + history（去门禁）。
3. Hub Ctrl+F5；确认无基础版主推项。
4. gateway 不必改 skip。

## Open Questions

- 库中已插入的 `predict_capability_basic`：Ensure `status=0` 即可，或运维手工下架（实现选 Ensure 置 0）。

## Why

先前 `predict-capability-unlock` 落地了「基础版 + Pro」双档，并用开通态门禁 imminent 推送。产品已改口：预测推送应对齐 3.1.9 **全面开放**；商业侧本轮只需要 **Pro 开通壳**（前端可试用/付费/邀请开通），Pro 具体能力留给后续提案。基础版与推送门禁必须回退，避免未购用户收不到预测提醒。

## What Changes

- **BREAKING（相对未发布的双档门禁实现）**：删除预测能力基础版（`predict_capability_basic`）种子、SKU、试用与相关文案；voice/history **不再**按开通态拦截 imminent / LA upsert。
- 预测推送（imminent 可见推送与统一 Live Activity 中的预测段）恢复为 **不依赖** 功能开通或 VIP（回到 3.1.9 行为）。
- **保留** 预测能力 Pro（`predict_capability_pro`）对人开通壳：支付、邀请、点试用（7 日起算）、catalog 展示；**本轮不实现任何 Pro 业务能力**（服务端不得以 Pro 权益驱动推送或其它功能）。
- **保留** Pro 开通前置：独立喂养场景 `predict_pro_entry` + App eligibility API（**计入 usage**）。
- **保留** 功能开通域广告中档清理；**保留** catalog 返回 `status`（含下架项，前端自决展示）。
- 活跃变更 `predict-capability-unlock` 的双档/门禁语义以本变更为准；实施时回退其门禁与基础版代码，收敛为 Pro 壳。

## Capabilities

### New Capabilities

- `predict-pro-unlock-shell`：仅 Pro 商业开通壳（种子、试用/支付/邀请、喂养前置、无业务挂载）；明确推送不读 Pro 权益。
- `predict-push-ungate`：回退/禁止预测推送能力门禁，恢复 3.1.9 全员可推。

### Modified Capabilities

- `predict-capability-unlock`：（若仍作基线引用）删除基础版与「Pro⊃推送」要求；改为仅 Pro 壳。本变更以新 capability 为准，并对旧实现做代码回退。
- `predict-imminent-capability-gate`：撤销门禁要求（改为 MUST NOT 按开通拦截写入型 imminent）。
- `predict-pro-feeding-eligibility`：保留场景与 App API、开通前门禁；与「无 Pro 业务能力」并存。
- `feature-ad-unlock-remove`：保持广告清理（无回退）。
- `feature-catalog-entitlement`：保持 catalog `status` / 下架仍返回（无回退）。

## Impact

- **进程**：`cash-service`（删基础版种子分支、保留 Pro/eligibility/试用）；`voice-service` / `history-service`（移除 access 门禁与 strip）；`clients/cash`（可删 RemotePredictCapabilityAccess）。
- **gateway-app**：eligibility 仍计入 usage（勿加 skip）；catalog 仍 skip。
- **Hub**：仅 Pro 平铺项（无基础版）；喂养页保留 `predict_pro_entry`；广告勾选保持已清理态。
- **后续**：另开提案挂 Pro 实际能力时再读 `predict_capability_pro` 权益。

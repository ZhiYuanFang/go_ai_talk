## Why

预测推送（imminent）目前无商业门禁，无法按账号售卖「预测能力基础版 / Pro」。需要把开通走通：基础版门禁 imminent、Pro 独立喂养资格后开通（本轮不做 Pro 能力细节但 Pro 覆盖基础版推送），并清理功能开通域残留的广告开通逻辑；同时 catalog 需把功能下架状态交给前端展示。

## What Changes

- 新增两个账号维（对人）功能：`predict_capability_basic`（预测能力基础版）、`predict_capability_pro`（预测能力 Pro）；开通通道 `payment` + `invite_code`；Hub / App catalog 形态对齐成长轨迹。
- 两档均支持**显式点试用**：起算点为点击时刻，时长 **7 日**，资格按 `wx × feature`（同机试用作废）；试用写入账号维权益。
- VIP 覆盖**使用**（含 imminent），不写功能权益行；catalog `unlocked` 不因 VIP 改写。
- **无**基础版权益、**无** Pro 权益、**非** VIP 时，Go **不推送/不启动/不更新 imminent**（含 Imminent Live Activity 的 start/update）；**timing Live Activity 不挡**。Pro 能力包含基础版，故有 Pro 亦可推 imminent。
- **Live Activity 清屏（end / dismissal）不受预测能力开通门禁影响**，避免权益到期后锁屏实时活动常驻无法摘除。
- Pro 开通（试用 / 付费 / 邀请）**不依赖**基础版；前置为当前绑机连续有效喂养资格，新建独立 scene `predict_pro_entry`（种子默认 7 日 × 每日 ≥10 条，Hub 可单独调）；提供 App eligibility API，字段同构 UCG/值得留意。
- **Pro eligibility 计入 App usage 统计**（不得写入 `maintenance_skip`）。
- App 功能目录 **MUST 返回 `status`（含下架 0）**，下架项仍出现在列表，由前端决定展示；服务端对新开通/建单仍拒绝下架功能。已购使用默认**停售不停用**（机权有效仍可推）。
- 功能开通域广告开通清到中档：Hub 去广告勾选、拒绝 `unlock_methods` 含 `ad`、下线/拒绝 App ad complete 与 Activate ad 分支、catalog 不再依赖广告字段。
- 本轮 **MUST NOT** 实现 Pro 具体业务能力（除「覆盖基础版推送」外）；**MUST NOT** 复活 `prediction_unlock` 条数模型。

## Capabilities

### New Capabilities

- `predict-capability-unlock`：预测基础版/Pro 种子、对人开通、7 日点试用、VIP 使用覆盖、Hub/catalog 对齐成长轨迹、Pro⊃基础版推送语义。
- `predict-pro-feeding-eligibility`：独立 scene `predict_pro_entry`、App eligibility API（计入 usage）、Pro 开通前门禁。
- `predict-imminent-capability-gate`：imminent start/update/可见推送按「基础版∨Pro∨VIP」门禁；timing 不挡；**LA 清屏 end/dismiss 永不因未开通而跳过**。
- `feature-ad-unlock-remove`：功能开通域广告开通中档清理。

### Modified Capabilities

- `feature-catalog-entitlement`：catalog 返回含下架项及 `status`；不再「停用则从目录消失」。
- `feeding-eligibility-admin`：登记/种子场景增加 `predict_pro_entry`。
- `feature-unlock-fulfillment`：广告通道拒绝；预测两档 InviteOncePerUser / 点试用履约约定（与广告清理、试用扩展交叉处）。

## Impact

- **进程**：`cash-service`（schema/种子、catalog、试用、eligibility、履约）；`voice-service` / 推送路径（imminent 门禁，经 cash internal 或 clients）；`gateway-app-server`（反代已有 cash 前缀则可能无需新 Bind；**usage**：eligibility **计入**，勿加 skip；若新 App 路径需确认 auth 豁免——默认须登录）。
- **Hub**：`cash-feature-admin.html`（去广告、新功能平铺）；喂养资格门槛页展示新 scene。
- **App**：catalog `status`；点试用 API；Pro eligibility；开通中心展示两档。
- **配置/ENV**：无新增强依赖 env（喂养阈值走 DB 场景表）。
- **重部署（实施后）**：至少 `cash-service`、`voice-service`（及实际挂门禁的 history/push 若改到）；gateway 若未改 skip/反代可不必，但若登记新路由 summary 建议随发版。

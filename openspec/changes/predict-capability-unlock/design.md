## Context

商业功能开通已在 cash-service：`feature_def` / `feature_user_entitlement` / catalog / 支付邀请履约；成长轨迹与值得留意均为对人开通，试用为账号维 `feature_user_trial`（24h、成功后 claim）。广告开通 App 路径已 stub，但 Hub/DTO/Activate 仍残留。`prediction_unlock` 条数模型已下线。imminent 推送（voice → push / Live Activity）当前无功能权益门禁；timing Live Activity 独立路径。喂养资格引擎已支持多 scene（`ucg_entry`、`care_alert_entry`），Hub 可改阈值。

## Goals / Non-Goals

**Goals:**

- 种子并对人开通「预测能力基础版 / Pro」；点试用 7 日起算；VIP 覆盖使用。
- Pro 开通前门禁：独立喂养 scene + App eligibility（**计入 usage**）。
- imminent 按「基础版∨Pro∨VIP」门禁；timing 不挡；Pro⊃基础版推送。
- **LA 清屏（end/dismissal）不受开通门禁**，防止到期后常驻。
- catalog 返回 `status`（含下架项）；广告开通中档清理。

**Non-Goals:**

- Pro 具体能力实现（除覆盖基础版推送外）。
- 复活 `prediction_unlock` / allowed_count。
- 同机试用同步；对机权益主体。
- 下架后强制停用已购推送（默认停售不停用）。
- 删除 `ad_duration_days` 列或历史订单迁移。
- 改变 UCG / care-alert eligibility 的 usage skip 策略。
- 用开通门禁去「惩罚性」拦截 end（明确禁止）。

## Decisions

### D1: feature_id 与主体

- `predict_capability_basic`、`predict_capability_pro`；`activation_subject=user`。
- 种子 `unlock_methods=payment,invite_code`；各一颗 30 天 SKU（`feat_predict_basic_30d` / `feat_predict_pro_30d`，价格占位 INSERT IGNORE，运维可改）。
- InviteOncePerUser 对两档生效（与 care/growth 一致）。
- **备选（否决）**：对机主体——已改口对人，且同机试用作废。

### D2: 点试用 7 日（人对人）

- App 显式领取（如 `POST /cash/app/api/feature/trial/start`），入参 `featureId`；主体 wx 取鉴权头。
- 起算：`now`；`expires_at = now + 7*86400`；`unlock_method=trial`；`feature_user_trial` unused→used。
- 已有有效人权益或 VIP：幂等跳过 claim（对齐成长轨迹「VIP 不消耗试用」语义）；已 used 拒绝再试。
- 成长轨迹 / 值得留意仍走原 24h soft→成功后 internal claim，**本变更不改其时长与触发**。
- 预测试用 Duration 用天（7），Activate 已支持 DurationHours；实现可用 DurationDays=7 或 hours=168。

### D3: Pro 喂养资格

- 新 scene `predict_pro_entry`，种子 7/10，INSERT IGNORE；Hub Admin 列表自动可见。
- App：`GET /cash/app/api/predict-pro/eligibility`，响应同构 UCG（qualified / requiredDays / effectiveDays / remainingDays / message）。
- **计入 usage**：不得加入 `maintenance_skip.go`；须有 `g.Meta` summary 供 apiregistry。
- Pro 试用 / 建单 / 邀请兑换前 MUST `qualified`；基础版开通不过此门。
- 资格按 **当前绑机 deviceNo** 的喂养数据计算（与现网引擎一致）。

### D4: imminent 门禁与 LA 清屏豁免

- **门禁范围（须校验开通）**：imminent 普通可见推送；Imminent / 统一投影路径上的 Live Activity **start** 与 **update**（向锁屏写入或刷新预测相关内容）。
- **门禁判定**（每个目标 wx）：`HasActiveUserFeature(basic) ∨ HasActiveUserFeature(pro) ∨ isVip`。
- **豁免（禁止门禁）**：
  - **Live Activity end / dismissal（清屏）**——无论因投影为空、权益到期后刷新、或显式 end；未开通/已过期也 MUST 下发 end，避免常驻。
  - **timing** Live Activity 全路径（start/update/end）不调用预测能力门禁。
- 权威在 cash：internal access；voice/history 经 `clients/cash`，禁止直查 cash 库。
- 实现注意：统一 `RefreshBabyLiveActivity` 若先判无权益就整段 return，会误伤 end；应按「有内容需 start/update → 校验；仅 end/清屏 → 放行」分支。
- 功能 def `status=0`：不影响已购推送（停售不停用）。

### D5: catalog `status` 与下架

- `ListActiveFeatureDefs` 改为列出需下发 App 的定义：**含 status=0**（仍排除硬编码已废 `prediction_unlock`）。
- 每项返回 `status`（0|1）；`unlocked` / `products` 语义：products 仍仅 status=1 SKU；下架功能 products 可为空。
- 建单 / Activate / 点试用：feature 或目标 SKU 非上架则拒绝。
- **BREAKING（相对旧规格）**：停用项不再从目录消失。

### D6: 广告中档清理

- Hub 去掉 ad 勾选与「邀请/广告天数」合并文案改为仅邀请天数。
- Admin 保存：`unlock_methods` 含 `ad` 拒绝；`adDurationDays` 校验可放宽或忽略写入。
- `CompleteFeatureAd` 保持拒绝；Activate `channel=ad` 拒绝。
- catalog 项不再返回 `adDurationDays`（或恒 0 且文档标明废弃——优先省略字段，注意旧客户端兼容：若需兼容可保留键恒 0，design 选 **省略并接受旧端忽略**）。

### D7: 内部与包边界

- cash 实现；voice/history 仅 clients；Redis 门禁缓存若需要走 cachekit + keys_* 登记（可先直查 entitlement 热路径，与 growth access 同级，是否缓存另评；**默认不新增 Redis 读缓存**，对齐「不默认引入 Redis」）。

## Risks / Trade-offs

- [catalog 含下架项] → 旧客户端若假设「列表即全部可买」可能展示下架功能 → 前端用 `status` 处理；服务端硬拦开通。
- [VIP 用户无权益行仍可收 imminent] → 与成长轨迹一致；catalog 仍显示未开通。
- [Pro eligibility 计入 usage] → 与 UCG/care eligibility「不统计」不一致 → 产品已确认，tasks 标明勿 skip。
- [点试用与 soft claim 两套] → 代码分支按 featureId 区分，避免误伤 care/growth。
- [imminent 每 wx 调 cash] → fanout 放大 → 可批量或短缓存后续优化；本轮正确性优先。
- [门禁误伤 LA end] → 到期常驻锁屏 → 规格强制清屏豁免；实现按 start/update vs end 分支，Refresh 禁止无权益整段短路。

## Migration Plan

1. 部署 cash-service（EnsureSchema 种子功能 + scene + SKU）。
2. 部署 voice-service（imminent 门禁）。
3. gateway-app：确认新路径反代与 usage（eligibility **不** skip）；Ctrl+F5 Hub。
4. App 发版配合 catalog `status`、点试用、Pro eligibility。
5. 回滚：门禁开关或快速放行（可选配置 `predictCapabilityGateEnabled`，默认 true）——若加开关须在 tasks 写明；**默认无开关、随版本回滚**。

## Open Questions

- 种子 SKU 占位价具体分值（实现可用 990 / 1900 对齐 care/growth，运维可改）。
- 点试用 API 最终 path 以 api/v1 Meta 为准（建议挂 feature 域）。

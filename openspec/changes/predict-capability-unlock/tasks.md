## 1. Schema 与种子

- [x] 1.1 在 `feature_const.go` 增加 `predict_capability_basic` / `predict_capability_pro` 常量与 SKU 编码；`InviteOncePerUser` 覆盖两档
- [x] 1.2 EnsureSchema 种子两档 `feature_def`（user、payment/invite_code、无 ad）及 30 天 SKU（INSERT IGNORE）
- [x] 1.3 EnsureSchema 种子 `feeding_eligibility_scene.predict_pro_entry`（7/10，INSERT IGNORE）；`defaultScene` / Admin 登记场景含该 key

## 2. 点试用（预测 7 日）

- [x] 2.1 实现 App 点试用 API（如 `POST /cash/app/api/feature/trial/start`）：仅预测两档走 7 日起算；写 `feature_user_trial` + user entitlement；VIP/已开通不消耗
- [x] 2.2 控制器 + `api/v1` Meta；确认须登录、计入 usage（非维护型）；gateway 反代/白名单按既有 cash feature 路径检查
- [x] 2.3 catalog `trialAvailable` 对预测两档按人维试用未用 ∧ 无有效人权益合成（同机不同步）

## 3. Pro 喂养资格

- [x] 3.1 实现 `GetPredictProFeedingEligibility`（复用 `GetFeedingEligibilityByScene`）
- [x] 3.2 App `GET /cash/app/api/predict-pro/eligibility` 字段同构 UCG；**不得**写入 `maintenance_skip.go`（负责人确认计入 usage）
- [x] 3.3 Pro 点试用 / 建单 / 邀请兑换前校验 `predict_pro_entry.qualified`；基础版路径不校验

## 4. Catalog 下架 status 与广告字段

- [x] 4.1 catalog 改为返回含 `status=0` 的功能项（仍排除 `prediction_unlock`）；每项输出 `status`
- [x] 4.2 下架功能拒绝建单/点试用/邀请开通；已购 unlocked 仍可真实反映
- [x] 4.3 catalog 去掉对 `adDurationDays` 的依赖（省略字段或废弃）；刷新 defs 缓存键语义

## 5. 广告开通中档清理

- [x] 5.1 ActivateFeature `channel=ad` 拒绝；保持/强化 `CompleteFeatureAd` 错误
- [x] 5.2 Admin 更新 def：拒绝 unlock_methods 含 `ad`；邀请天数校验保留 ≥1；Hub 去掉广告勾选与广告天数文案
- [x] 5.3 care/growth/预测 Ensure 强制 `unlock_methods` 不含 ad

## 6. imminent 门禁与 LA 清屏豁免

- [x] 6.1 cash internal access API：按 wx 合成 basic∨pro∨VIP（可选分项字段）
- [x] 6.2 `clients/cash` 封装；imminent **可见推送**与 LA **start/update** 前按目标 wx 校验，失败 skip 并打日志
- [x] 6.3 **LA end/dismissal（清屏）不校验开通**；统一 Refresh 禁止「无权益整段 return」误伤 end；验证到期后可清屏
- [x] 6.4 确认 timing Live Activity 全路径不调用预测能力门禁

## 7. Hub 与规则文案

- [x] 7.1 `cash-feature-admin.html` 平铺含预测两档；规则摘要对人 + Pro 喂养前置说明
- [x] 7.2 喂养资格 Hub 页可见并编辑 `predict_pro_entry`

## 8. 收尾校验

- [x] 8.1 跑 `hack/check-service-import`（或 ps1）与 Redis bypass 检查相关改动
- [x] 8.2 自检：新 App 路径 Meta/summary、usage（eligibility 计入、catalog 仍 skip）、无 Bearer 豁免误加
- [x] 8.3 回复中列出重部署进程：`cash-service`、`voice-service`；gateway-app 若未改 skip/反代可注明；Hub 静态需刷新

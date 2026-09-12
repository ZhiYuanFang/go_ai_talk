## 1. Schema 与常量

- [x] 1.1 EnsureSchema：`feature_def` 增加 `activation_subject`（VARCHAR，默认 `device`）；已有行回填 `device`；种子成长轨迹设 `user`
- [x] 1.2 EnsureSchema：新建 `feature_user_entitlement`（`wx_id`+`feature_id` 唯一，字段镜像设备权益表语义）
- [x] 1.3 常量：`InviteOncePerUser`（至少含 `growth_trajectory_predict`）；注释标明与 `InviteOncePerDevice` 并存

## 2. Activate / Grant 双路径

- [x] 2.1 `ActivateFeature` 支持 `ActivationSubjectUser`：解析 SubjectKey 为 wxId；预测/条数类拒绝 user
- [x] 2.2 实现 user 权益 upsert/续期与 `HasActiveUserFeatureEntitlement`；失效相关缓存
- [x] 2.3 支付履约：按 `feature_def.activation_subject` 选择 deviceNo 或 order.wxId 作为 SubjectKey
- [x] 2.4 邀请兑换：按 subject 调用 Activate；成长轨迹写 user；保留设备邀请闸；新增人×功能一次闸
- [x] 2.5 广告完成：user 主体要求 wxId；按 subject 授予；幂等策略与注释对齐 design

## 3. Access / Catalog / 成长轨迹门禁

- [x] 3.1 `GetGrowthTrajectoryAccess` 仅认 user 权益 ∨ VIP（忽略旧 device 行）
- [x] 3.2 `GetFeatureCatalog` 注入 wxId：device/user 功能分别读对应权益表
- [x] 3.3 App/controller 路径确保 catalog 传入登录 wxId
- [x] 3.4 voice `GrowthTrajectoryLatest`：保留登录校验，移除开通/VIP 校验
- [x] 3.5 voice 日限：cachekit 键改 wxId；turn 预检/INCR/single-flight 跟人

## 4. Admin API 与后台页

- [x] 4.1 Admin defs GET/POST 增加 `activationSubject`；预测改 user 拒绝；未传保持原值
- [x] 4.2 `cash-feature-admin.html` 列表展示「对人/对机」；编辑区展示并可保存主体
- [x] 4.3 Hub 入口无缓存刷新提示（既有 Ctrl+F5 文案可沿用）

## 5. 验收与文档

- [x] 5.1 自检：A 开通后仅 A 可 turn；同设备 B 非 VIP 不可 turn；双方登录可读 latest（代码路径：access=user∨VIP；latest 无开通闸）
- [x] 5.2 自检：邀请同人二次失败、同设备第二人失败；日限同人跨设备共享（代码路径：InviteOncePerUser+Device；日限键 wxId）
- [x] 5.3 自检：Admin 列表成长轨迹「对人」、值得留意「对机」（页面 subjectLabel + API 字段）
- [x] 5.4 （可选）runbook 一句：成长轨迹无手工 DDL；EnsureSchema + 不迁存量 device 行

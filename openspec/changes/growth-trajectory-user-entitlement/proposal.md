## Why

成长轨迹当前按 `device_no` 全家共享开通，与业务「一人一份权限（跟 wx.id）」不符；读最新报告仍校验开通，也与「仅登录即可按设备读历史」不符。同时开通原子仅支持设备主体，无法干净落地账号维权益；后台功能管理也看不出各功能是对人还是对机，后续维护易混。

## What Changes

- **通用化 Activate**：真正落地 `ActivationSubjectUser`；`feature_def` 增加 `activation_subject`（`device`|`user`），支付/邀请/广告按定义选权益主体。
- **账号维权益表**：新增 `feature_user_entitlement`（`wx_id`+`feature_id`）；设备表 `feature_entitlement` 语义不变。
- **成长轨迹切 user**：种子 `growth_trajectory_predict` 的 `activation_subject=user`；付费仍 19 元/30 天非永久；VIP 仍可免开通（与 user 权益 OR）。
- **邀请双闸**：成长轨迹保留设备邀请一次（`InviteOncePerDevice`），并新增「同一人对该功能邀请仅一次」（`InviteOncePerUser`）。
- **日限跟人**：成长轨迹每日次数 Redis 计数由 `deviceNo` 改为 `wxId`。
- **读历史免开通**：`GET .../growth-trajectory/latest` 按 `deviceNo` 读最新一条，须登录、不校验开通/VIP；`turn` 仍须 user 开通 ∨ VIP。
- **后台功能管理**：列表与编辑区明确展示/可维护「开放能力主体：对人（user）/对机（device）」，便于运维追踪。
- 存量：若已有成长轨迹 **device** 权益行，本变更默认**不自动迁移**（忽略；冷启动按新规则写 user 表）。未上量环境可接受。

## Capabilities

### New Capabilities

- `feature-activation-subject`：Activate/履约按 `activation_subject` 写入 device 或 user 权益；catalog/access 按主体读取。
- `growth-trajectory-user-gates`：成长轨迹开通跟人、latest 免开通仍登录、turn 门禁、日限跟人、邀请人/设备双闸。
- `feature-admin-activation-subject`：Admin API 与「开通功能管理」页展示/更新对人|对机主体。

### Modified Capabilities

- （无独立主库 capability 名；行为增量见上方新 capability。既有 commercial / care-alert 变更中的设备维语义对值得留意等 **保持不变**。）

## Impact

- **cash-service**：`EnsureSchema`、`ActivateFeature`、`feature_grant`、支付履约、邀请兑换、广告完成、`GetGrowthTrajectoryAccess`、`GetFeatureCatalog`、Admin defs API。
- **voice-service**：`GrowthTrajectoryLatest` 去开通校验；日限 key 与 turn 计数改 wx；access 仍调 cash（turn）。
- **gateway-app**：catalog 若需 wx 合成，沿用已有登录头；Admin 静态页 `cash-feature-admin.html`。
- **库**：`ai_voice_cash` 新表/新列（EnsureSchema）；voice 库日限仅 Redis，无新表。
- **客户端**：catalog「成长轨迹是否开通」改跟当前登录账号；latest 未开通也可读到历史；换号不共享开通。

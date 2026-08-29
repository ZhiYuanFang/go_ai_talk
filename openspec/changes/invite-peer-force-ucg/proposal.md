## Why

邀请码仍按「运营营销激活码」建模（Admin 造码、码绑功能/有效期、兑码预测临时全开），与「用户互推、永久累加预测条数、获客激励原力」产品目标冲突；原力寄存在 device.`wx` 也不符合 UCG 域边界。功能尚未发布，适合一次性重写契约。

## What Changes

- **BREAKING**：邀请码改为 cash 侧用户身份凭证——注册/Ensure 自动一用户一码；废除码级有效期、功能白名单、`grant_duration`、一家锁定、Admin 造码页（整页删除）。
- **BREAKING**：预测开通统一为永久条数累加：支付 / 邀请码 / 广告 **每次 +1**；废除邀请兑码 T1 全开（`allowedCount=-1`）路径。
- 兑码去重改为 **人 × 码 × 功能**；允许同一用户兑多个好友码；允许 A↔B 对同一功能各兑一次。
- 获客数量（`redeemed_count`）每次成功兑 +1；获客 ×100 写入 **ucg 原力**（含流水 reason）；辩论自投继续 +1 并写流水；**不回填**历史投票。
- **BREAKING**：原力从 device.`wx.force_value` **迁至 ucg 本域**（计数 + ledger）；device 内部 force API / BatchWx 中的 force 字段移除。
- App：我的邀请码、邀请列表（昵称+时间）；原力当前值 + 积分流水（详情页由客户端算档位/距升级）。
- catalog 预测项聚合 **`totalActivatableCount`**：由 **device 事件字典统计非叶子事件数**，cash 读契约后写入 catalog（方案 C）。
- VIP 与单事件开通解耦：VIP 仅有效期内覆盖使用权，**不**改永久 `allowed_count` / entitlement；契约与文档明确此规则。
- 非预测功能若支持 `invite_code`：默认 grant 为 **权益一次**（entitlement）。
- 孪生客户端变更：`flutter_ai_talk` 同名 change `invite-peer-force-ucg`。

## Capabilities

### New Capabilities

- `invite-code-identity`：Ensure/注册发码、App 我的码与邀请列表、兑码规则（多码可兑、人×码×功能、无码级范围/有效期）。
- `prediction-count-grant`：支付/邀请/广告预测开通均为永久 +1；SKU `grantQuantity=1`；废 T1 全开。
- `ucg-force-ledger`：原力迁 ucg、ledger（`debate_self_vote` / `invite_acquisition`）、cash→ucg 获客加分契约、App 流水读接口。
- `catalog-activatable-total`：device 非叶子事件计数契约；cash catalog 聚合 `totalActivatableCount`。
- `vip-overlay-entitlement`：VIP 时效覆盖与永久单事件开通解耦的服务端语义（读模型叠加、写路径不合并）。

### Modified Capabilities

- （无已归档基线 capability 目录；行为以本 change 下 specs 为准，并废止未发布的营销邀请码/T1 全开约定。）

## Impact

- **cash-service**：邀请 Ensure/Redeem/App 读接口；删除邀请 Admin 静态页与 Hub 模块及造码 CRUD；履约与广告 +1；catalog 聚合 total；兑成功调 ucg 加分。
- **device-service**：注册后触发 Ensure（或等价回调）；事件非叶子计数内部/可聚合 API；移除 `wx.force_value` 与 force increment。
- **ucg-service**：`ucg_user_force` + `ucg_force_ledger`；投票写本域；profile enrich 改读本域；内部获客加分 API；App 流水接口。
- **gateway-app**：路由/白名单/usage 策略（新增 App 接口须询问 usage）；静态页拆除。
- **Flutter**：同名 change 个人中心 + 开通中心 UX。
- 无历史数据迁移；不新增测试文件。

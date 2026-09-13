## 1. 原力首次加分与资料展示

- [x] 1.1 修复 `AddForceDelta`：空 `ucg_user_force` 行可 INSERT，禁止空集 `Scan` 直接失败
- [x] 1.2 `profile/me`（`mergeProfileForAuthor` 或等价路径）调用 `enrichProfileForceValues`
- [x] 1.3 （可选）提供幂等历史获客补偿说明或脚本：ledger 无 `invite_acquisition`+code ref 时按 redemption 补 +100

## 2. 成长轨迹邀请闸

- [x] 2.1 `InviteOncePerDevice` 仅保留值得留意；成长轨迹不再校验/写入 `feature_invite_device_grant`
- [x] 2.2 确认 `InviteOncePerUser`、人×码×功能、自用、同宝宝禁兑仍对成长轨迹生效；更新相关中文注释

## 3. Admin 只读规则

- [x] 3.1 cash 增加按 `featureId` 派生的只读 `ruleSummary`（或常量映射），Defs 列表/详情 API 带出
- [x] 3.2 `cash-feature-admin.html` 展示只读规则区（含三功能口径与同宝宝禁兑），不可编辑保存

## 4. Admin 开通快照（方案 A）

- [x] 4.1 新增 Admin API：按 `featureId` 分页返回当前开通快照（device/user/预测条数分支）
- [x] 4.2 `api/v1` 增加请求结构；controller 挂 Admin 鉴权；中文注释
- [x] 4.3 Admin 页：功能列表入口进详情，展示规则 + 快照表（有效/到期/方式；页内注明无 VIP、非事件历史）
- [x] 4.4 用户维快照尽力补昵称（ucg batch）；失败可空

## 5. 校验

- [x] 5.1 自检：首次获客加分成功；成长轨迹同机第二人不同码可兑、同一人二次拒；值得留意同机仍一次
- [x] 5.2 `openspec validate feature-ops-invite-force-admin --strict`

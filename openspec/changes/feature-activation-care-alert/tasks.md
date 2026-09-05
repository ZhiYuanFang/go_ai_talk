## 1. Schema 与功能种子

- [x] 1.1 EnsureSchema：种子 `care_alert_smart_remind`（标题/简介、unlock_methods 含 payment,invite_code,ad、duration_days 默认 7）；幂等且不覆盖运维已改值
- [x] 1.2 EnsureSchema：设备维邀请去重表（如 `feature_invite_device_grant`，PK device_no+feature_id）；中文注释标明适用功能
- [x] 1.3 种子值得留意付费 SKU（entitlement、duration_days=0、grant_quantity=1）；常量 `FeatureIDCareAlertSmartRemind`

## 2. 开通原子工具

- [x] 2.1 实现 `ActivateFeature`（或等价）：SubjectType/Channel 枚举、效果解析（payment←SKU；invite/ad←feature_def.duration_days；预测 count+1）
- [x] 2.2 支付履约、`RedeemInviteCode`、`CompleteFeatureAd` 改汇入原子入口；去掉旁路写表
- [x] 2.3 邀请路径：`care_alert_smart_remind` 校验 device×feature 仅一次；成功写去重表；原力仍 NotifyUcg 用户侧
- [x] 2.4 自检：预测邀请/广告/支付仍为永久 +1；同设备值得留意二次邀请拒绝

## 3. Admin

- [x] 3.1 Admin API/列表返回功能定义邀请/广告授予天数
- [x] 3.2 `cash-feature-admin.html` 功能定义表单增加天数字段与说明（0=永久；付费看套餐；预测忽略）

## 4. cash access 契约

- [x] 4.1 `GET /cash/internal/api/care-alert/access`：合成 feedingQualified、featureActive（权益∨VIP）、allowed；internal secret
- [x] 4.2 `api/v1` g.Meta + controller；Remote 客户端封装供 voice（对齐 RemoteIsVip）
- [x] 4.3 确认 VIP 失败当作非 VIP；整体失败由调用方 fail-closed

## 5. voice 双门禁

- [x] 5.1 `CareAlertDaily`（含 force 与缓存命中）调用 access；未 allowed 拒绝且不调 Python
- [x] 5.2 Delete/Feedback 同样 access 门禁
- [x] 5.3 voice 配置/文档确认 `CASH_SERVICE_URL` 可达

## 6. 网关与观测

- [x] 6.1 确认 internal/App 反代覆盖；不加入 Bearer 匿名白名单
- [x] 6.2 若新增 App 只读合成接口：先问负责人是否计入 usage；未答复不改 maintenance_skip
- [x] 6.3 中文注释覆盖新模块/关键分支

## 7. Flutter（flutter_ai_talk）

- [x] 7.1 常量 `care_alert_smart_remind`；目录/开通中心展示与 VIP OR 语义
- [x] 7.2 值得留意三态：未喂养合格进度；合格未开通可点进开通中心；合格且开通/VIP 拉 daily
- [x] 7.3 与 Go 联调：未开通 daily 被拒；VIP/付费/邀请到期行为符合预期

## 8. 收尾自检

- [x] 8.1 `rg`：voice 无直查 cash 库；开通写路径均经 Activate
- [x] 8.2 喂养资格仍不受 VIP/开通短路
- [x] 8.3 不新增 `*_test.go`；Ctrl+F5 Admin 页字段可见

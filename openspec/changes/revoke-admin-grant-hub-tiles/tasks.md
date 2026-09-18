## 1. 撤销服务

- [x] 1.1 增加订单状态 `revoked`（不用 `refunded`）
- [x] 1.2 实现 VIP 撤销：校验最近 paid 为 `channel=admin` 且未过期；`expire_at=now`；该订单标 `revoked`
- [x] 1.3 实现功能撤销：按 `activation_subject` 校验主体；同样立即过期并标 `revoked`；拒绝 `prediction_unlock`
- [x] 1.4 `api/v1` + controller：`POST /cash/admin/api/vip/entitlements/revoke` 与 `POST /cash/admin/api/feature/grants/revoke`（Admin 口令）

## 2. 去掉预测槽位

- [x] 2.1 `EnsureSchema` 不再 INSERT/UPDATE `prediction_unlock`
- [x] 2.2 删除履约、下单、Apple 通知、catalog、Admin 规则/快照中的预测特判；遇到该 `featureId` 拒绝并打日志
- [x] 2.3 删除不再被调用的 `allowed_count_delta` 分支与常量；不 DROP `feature_allowed_count`

## 3. Hub 平铺

- [x] 3.1 `cash-feature-admin.html` 顶部平铺 VIP、上架功能、群二维码；点选只渲染当前面板
- [x] 3.2 VIP 面板收纳套餐、权益列表、授表；`channel=admin` 且有效的行可撤销；`cash-vip-admin.html` 不再作为第二套可写入口
- [x] 3.3 功能面板固定当前 `featureId` 的定义/SKU/快照/授撤；去掉预测条数与 `allowed_count_delta` 表单项
- [x] 3.4 群二维码仅在自身面板编辑

## 4. 自检

- [x] 4.1 确认未改 App 支付契约、未改 `maintenance_skip`、未引入按天数回退或自动退款
- [x] 4.2 `openspec validate revoke-admin-grant-hub-tiles --strict` 通过

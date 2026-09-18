## ADDED Requirements

### Requirement: 系统 MUST NOT 再种子或特判 prediction_unlock

cash-service `EnsureSchema` MUST NOT 再插入或每次启动写回 `feature_def.feature_id=prediction_unlock`。履约、下单、Apple 通知、Admin 规则与 catalog MUST NOT 再为该 `feature_id` 走「预测条数 / `allowed_count_delta`」专用分支。若请求仍携带该 `feature_id`，系统 MUST 拒绝开通并记录日志，MUST NOT 把它当成普通权益授予。

#### Scenario: 启动不再写预测种子

- **WHEN** cash-service 执行 EnsureSchema
- **THEN** 系统 MUST NOT 执行针对 `prediction_unlock` 的 INSERT 或将其 `status` 写回的 UPDATE

#### Scenario: 旧请求被拒绝

- **WHEN** 支付履约或 Admin 授功能的 `featureId` 为 `prediction_unlock`
- **THEN** 系统 MUST 拒绝，MUST NOT 增加 `feature_allowed_count`

### Requirement: Admin MUST NOT 再提供预测条数表单项

`cash-feature-admin.html` MUST NOT 再展示「增加可看数量」、默认开通条数，或手工授里的授予数量字段。SKU 的 `grant_kind` 表单 MUST NOT 再提供 `allowed_count_delta`。

#### Scenario: 打开功能 SKU 表单

- **WHEN** 运维编辑任一上架功能的套餐
- **THEN** 表单 MUST NOT 出现 `allowed_count_delta` 选项

### Requirement: 删除预测代码 MUST NOT 删除历史表

实现 MUST NOT DROP `feature_allowed_count`，MUST NOT 删除已有 `feature_order` / `feature_product` 历史行。其它功能（值得留意、成长轨迹、邀请码、群二维码、VIP）的履约路径 MUST 保持可用。

#### Scenario: 历史表仍在

- **WHEN** 本变更部署完成
- **THEN** `feature_allowed_count` 表 MUST 仍然存在，且 care / growth 的支付与邀请开通 MUST 仍可成功

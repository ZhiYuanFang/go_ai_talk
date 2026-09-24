## ADDED Requirements

### Requirement: HMS 推送 MUST 按 bizType 设置 android.notification.category

当 push-service 经 HMS 通道发送推送时，若请求载荷中的业务类型 `bizType`（`payload.Data["bizType"]`，大小写不敏感比较前 MUST 规范化为小写 trim）为已知可见类型，则 MUST 在 `message.android.notification` 中设置 `category` 为华为官方大写枚举；MUST NOT 将 `category` 写入顶层 `message.notification` 以外的错误位置（本需求仅约束 android.notification）。

映射 MUST 为：

| bizType（规范化后） | category |
|---------------------|----------|
| `predict_imminent`  | `WORK`   |
| `ucg_alert`         | `IM`     |

#### Scenario: 预测临近推送带 WORK

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `predict_imminent`
- **THEN** 请求体中 `message.android.notification.category` MUST 为字符串 `WORK`

#### Scenario: UCG 可见内容推送带 IM

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `ucg_alert`
- **THEN** 请求体中 `message.android.notification.category` MUST 为字符串 `IM`

### Requirement: HMS 静默角标与未知 bizType MUST NOT 写入 category

当规范化后的 `bizType` 为 `ucg_silent_badge`，或为空，或为上述映射表以外的任意值时，HMS 请求体中的 `message.android.notification` MUST NOT 包含 `category` 字段（键不存在，而非空字符串）。

#### Scenario: UCG 静默角标不传 category

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `ucg_silent_badge`
- **THEN** `message.android.notification` MUST NOT 含有 `category` 键

#### Scenario: 未知或空 bizType 不传 category

- **WHEN** HMS 发送且 `bizType` 为空或未落入 `predict_imminent` / `ucg_alert` / `ucg_silent_badge` 的已知映射
- **THEN** `message.android.notification` MUST NOT 含有 `category` 键

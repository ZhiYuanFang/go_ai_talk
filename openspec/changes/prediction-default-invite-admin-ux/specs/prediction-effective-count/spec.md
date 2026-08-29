## ADDED Requirements

### Requirement: 预测有效条数 MUST 为默认加永久累加

对 `prediction_unlock`（及约定的预测数量类功能），合成 catalog 暴露的有效可看条数（非临时全开时）MUST 等于 `defaultFree + permanentDelta`，其中 `defaultFree` 来自功能定义可配置默认开通数（非负），`permanentDelta` 来自设备永久累加权威（付费/广告等，非邀请码临时全开）。无设备永久行时 `permanentDelta` MUST 为 0。系统 MUST NOT 在无临时全开时仅因缺行而固定返回 0 并忽略已配置的 `defaultFree`。

#### Scenario: 仅配置默认时返回默认条数

- **WHEN** 某设备无永久累加记录，且该功能 `defaultFree=3`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `3`

#### Scenario: 默认加永久累加

- **WHEN** `defaultFree=3` 且该设备 `permanentDelta=5`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `8`

### Requirement: 邀请码预测临时全开 MUST 使用哨兵 -1

当设备对预测功能存在有效临时全开（邀请码授予且未过期，或永久全开标志生效）时，catalog 该项 `allowedCount` MUST 为 `-1`（哨兵，表示全部可看）。有限期临时全开 MUST 暴露到期时间（`expiresAt` 为 unix 秒）。到期后 MUST 回落为「默认 + 永久累加」，且 MUST NOT 清除既有永久累加。邀请码兑换预测 MUST NOT 再将 `grant_quantity` 累加进永久 `allowed_count`。

#### Scenario: 期限内返回 -1

- **WHEN** 用户用邀请码成功开通 `prediction_unlock` 且授予有效天数 > 0，且当前时间未超过到期
- **THEN** catalog 该项 `allowedCount` MUST 为 `-1`，且 MUST 含对应 `expiresAt`

#### Scenario: 到期后回落

- **WHEN** 临时全开已过期，且 `defaultFree=2`、`permanentDelta=4`
- **THEN** catalog 该项 `allowedCount` MUST 为 `6`，且 MUST NOT 为 `-1`

#### Scenario: 邀请码不增加永久条数

- **WHEN** 邀请码兑换预测成功
- **THEN** 该设备永久累加权威值 MUST 与兑换前相同（仅更新临时全开状态）

### Requirement: 默认开通数 MUST 可经 Admin 配置

功能定义 MUST 支持配置预测类默认开通数；Admin 更新后，后续无临时全开的 catalog 读 MUST 反映新默认（缓存失效后）。功能编号仍为与客户端约定的稳定 ID，本要求不授权 Admin 修改 `featureId`。

#### Scenario: Admin 修改默认开通数

- **WHEN** 运维将 `prediction_unlock` 的默认开通数从 0 改为 2 并保存
- **THEN** 无永久累加且无临时全开的设备再次拉 catalog 时该项 `allowedCount` MUST 为 `2`

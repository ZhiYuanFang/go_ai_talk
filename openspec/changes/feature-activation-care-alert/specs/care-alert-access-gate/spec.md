## ADDED Requirements

### Requirement: cash-service MUST 提供值得留意可看合成内部契约

cash-service MUST 提供经内部密钥保护的只读接口（路径以实现为准，如 `GET /cash/internal/api/care-alert/access`），入参 MUST 含 `deviceNo` 与 `wxId`。响应 MUST 至少包含：`allowed`（bool）、`feedingQualified`（bool）、`featureActive`（bool）。合成规则 MUST 为：`feedingQualified` 取自与 App `care-alert/eligibility` 同构的 `care_alert_entry` 判定；`featureActive` 为该设备 `care_alert_smart_remind` 未过期权益 **或** 该 `wxId` 当前 VIP；`allowed = feedingQualified AND featureActive`。VIP / 功能开通 MUST NOT 将 `feedingQualified` 置 true。voice-service MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 双门槛均满足

- **WHEN** 设备喂养资格合格且（有未过期值得留意权益或账号 VIP）
- **THEN** 响应 MUST `allowed=true`、`feedingQualified=true`、`featureActive=true`

#### Scenario: 仅喂养合格未开通非 VIP

- **WHEN** `feedingQualified=true` 且无有效权益且非 VIP
- **THEN** 响应 MUST `allowed=false`、`featureActive=false`

#### Scenario: 已开通但喂养不合格

- **WHEN** 设备有永久值得留意权益但喂养资格不合格
- **THEN** 响应 MUST `allowed=false`、`feedingQualified=false`、`featureActive=true`

#### Scenario: VIP 覆盖开通但不覆盖喂养

- **WHEN** 用户 VIP 有效、无设备权益、喂养不合格
- **THEN** 响应 MUST `feedingQualified=false`、`featureActive=true`、`allowed=false`

### Requirement: voice CareAlert 业务接口 MUST 经 access 双门禁

voice-service 在处理值得留意日列表（`CareAlertDaily`，含强制刷新与缓存命中路径）时 MUST 先调用 cash 值得留意 access 契约；仅当 `allowed=true` 时 MAY 返回缓存内容或触发生成。当 cash 不可达、超时或返回错误时，系统 MUST fail-closed 拒绝该次业务（MUST NOT 当作已开通放行），并记录 Warning。删除单条与反馈接口 MUST 同样调用 access 且仅在 `allowed=true` 时执行变更（与 Daily 一致，防止过期后改缓存）。

#### Scenario: 未开通拒绝 daily

- **WHEN** access 返回 `allowed=false`（例如合格未开通）
- **THEN** CareAlertDaily MUST 返回错误或空拒绝语义，MUST NOT 返回日列表内容，MUST NOT 调用 Python 生成

#### Scenario: 缓存命中仍校验

- **WHEN** 当日日缓存存在但 access 返回 `allowed=false`（如邀请已过期）
- **THEN** 系统 MUST NOT 向客户端返回该缓存内容

#### Scenario: cash 故障 fail-closed

- **WHEN** voice 调用 cash access 失败
- **THEN** CareAlertDaily MUST 失败关闭，MUST NOT 降级为跳过开通检查

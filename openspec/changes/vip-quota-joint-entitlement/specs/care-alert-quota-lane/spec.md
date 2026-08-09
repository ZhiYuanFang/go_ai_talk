## ADDED Requirements

### Requirement: care-alert 独立额度 feature

voice-service MUST 提供独立 AI 额度 feature **`care_alert`**（与 `voice_ai` / `clinic_ai` 并列），含全局默认上限、per-wxId override、按月用量与 App/Admin 可读状态。care-alert 日生成路径 MUST 使用该 feature 参与统一 premium 判定，MUST NOT 再用 `clinic_ai` 代替 care-alert 额度，MUST NOT 仅根据 VIP 布尔硬切 DeepSeek/Zhipu。

#### Scenario: 非 VIP 有 care_alert 额度走正式模

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=true
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python 分析

#### Scenario: 非 VIP 且 care_alert 用尽走 free/omit

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=false
- **THEN** 服务 MUST 按 `careAlert` lane 的 free 配置选模；free 为空 MUST omit model 交 Python

#### Scenario: VIP 走正式模

- **WHEN** 日缓存未命中且触发者为 VIP
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python，MUST NOT 因 used 达旧 limit 改走 free

#### Scenario: Hub 可配置 care_alert 月度额度

- **WHEN** 管理员打开 `/device/admin/voice-admin.html` 或调用 voice Admin `ai-quota` API
- **THEN** 全局默认 MUST 可读/写 `careAlertMonthlyLimit`；用户列表 MUST 展示 `careAlert` 已用与上限，且 PUT 用户 override MUST 可提交 `careAlertMonthlyLimit`

### Requirement: care-alert 成功时非 VIP 计次

护理留意日分析在 Python 成功并写入（或准备写入）日缓存的成功路径上，若触发者 **非 VIP**，系统 MUST 对 `care_alert` 执行 consume；若触发者为 VIP，MUST NOT consume。缓存命中直接返回时 MUST NOT check/consume 额度，MUST NOT 重新选模。

#### Scenario: 非 VIP 首次生成成功扣次

- **WHEN** 非 VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST 增加 1（在 limit 允许的成功语义下）

#### Scenario: VIP 首次生成不计次

- **WHEN** VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST NOT 因本次成功增加

#### Scenario: 缓存命中不计次

- **WHEN** 当日缓存已存在
- **THEN** 服务 MUST 直接返回列表，MUST NOT consume `care_alert`

### Requirement: careAlert LLM lane 与并发

系统 MUST 提供独立 lane **`careAlert`**，具备与其它 voice lane 相同的 `provider`/`model`/`maxInFlight`/`maxWaiters` 及 free 字段，并经 aimodel 闸门 `Acquire` 控制并发（premium 或已配置 free 时按 design 约定 Acquire）。care-alert 生成 MUST 使用该 lane，MUST NOT 再借用 clinic 闸门数值并硬编码供应商切换。

#### Scenario: Admin 可配置 careAlert

- **WHEN** 管理员 GET/PUT `/voice/admin/api/llm-lanes`
- **THEN** 响应/请求 MUST 包含 `careAlert` 子对象（含正式模与 free 字段）

#### Scenario: 生成走 careAlert 闸门

- **WHEN** premium 路径调用 Python care-alert 分析
- **THEN** 服务 MUST 在调用前对该 lane 执行 Acquire（队列满语义与其它 lane 一致）

### Requirement: care-alert 仍要求有效 wxId

care-alert HTTP 接口 MUST 继续要求 `wxId>0`；MUST NOT 赋予硬件特权；MUST NOT 用 deviceNo 反查替代登录。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求无有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 拒绝，MUST NOT 生成或计次

## REMOVED Requirements

### Requirement: care-alert 仅按 VIP 硬切 DeepSeek/Zhipu

**Reason**: 由 VIP∪`care_alert` 额度统一权益与 `careAlert` lane 正式/free 配置取代。  
**Migration**: 实现与运维改为配置 `careAlert` lane；删除 `resolveCareAlertModelProfile` 内 DeepSeek/Zhipu 硬切逻辑；更新 `llm-care-alert-daily` CONTRACT 与相关 runbook 表述。

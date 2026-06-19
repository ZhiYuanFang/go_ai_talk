## ADDED Requirements

### Requirement: voice-service SHALL 提供 LLM lane Admin API

voice-service MUST 提供 `GET /voice/admin/api/llm-lanes` 与 `PUT /voice/admin/api/llm-lanes`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（经 gateway-app 反代）。响应与请求 MUST 包含 `voiceUnderstanding` 与 `clinic` 两个子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist 供 Admin UI 下拉联动。PUT MUST 校验 allowlist 与正整数边界（`maxInFlight>=1`，`maxWaiters>=0`）。

#### Scenario: 管理员读取 lane 配置

- **WHEN** 运维携带正确口令 GET `/voice/admin/api/llm-lanes`
- **THEN** 响应 SHALL 含两 lane 当前 DB 配置与 allowlist

#### Scenario: 管理员更新 clinic 并发

- **WHEN** 运维 PUT `clinic.maxInFlight=5` 且 model 在 allowlist 内
- **THEN** voice-service MUST 持久化至 `ai_voice_voice` 且 MUST 失效 lane 缓存

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 MUST 返回未授权且 MUST NOT 修改配置

### Requirement: ucg-service SHALL 扩展 AI 配置 Admin 以包含 polish lane 闸门

`GET/PUT /ucg/admin/api/ai-config` MUST 扩展支持 `provider`、`maxInFlight`、`maxWaiters`（与现有 `visionModel` / `maxImagesPerRequest` 并存或语义对齐为 polish lane 的 `model`）。PUT 成功后 MUST 失效 AI/lane 缓存（与现有 `InvalidateAIConfigCache` 一致）。

#### Scenario: 管理员更新润笔模型与缓冲池

- **WHEN** 运维 PUT `provider=zhipu`、`visionModel=glm-4.6v-flash`、`maxInFlight=1`、`maxWaiters=15`
- **THEN** ucg-service MUST 持久化且下一笔润笔 MUST 使用新 profile

### Requirement: DB 种子 MUST 默认为智谱三模型（方案 A）

首次 EnsureDefaultRows（或 migration 种子）时：若 lane 配置行不存在，系统 MUST 写入：`voiceUnderstanding` → `zhipu` / `glm-4.7-flash` / `maxInFlight=1` / `maxWaiters=20`；`clinic` → `zhipu` / `glm-4.1v-thinking-flash` / `maxInFlight=1` / `maxWaiters=10`；`polish` → `zhipu` / `glm-4.6v-flash` / `maxInFlight=1` / `maxWaiters=15`。YAML 中 DeepSeek/DashScope 默认值 MUST 保留作 DB 缺失时的冷启动兜底。

#### Scenario: 新环境首次启动

- **WHEN** voice 与 ucg 库尚无 lane 配置行
- **THEN** EnsureDefaultRows 后 Admin GET MUST 返回上述智谱默认

### Requirement: voice-admin 与 ucg-admin SHALL 提供 LLM lane 配置 UI

`resource/public/voice-admin.html` MUST 新增「LLM 车道」Tab，可编辑 `voiceUnderstanding` 与 `clinic` 的 provider、model、maxInFlight、maxWaiters。`resource/public/ucg-admin.html`「AI 配置」Tab MUST 扩展 provider、maxInFlight、maxWaiters 字段（model 沿用 visionModel 下拉或等价控件）。保存 MUST 调用对应 Admin PUT API。

#### Scenario: voice-admin 保存车道

- **WHEN** 运维在 voice-admin LLM 车道 Tab 提交修改
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes`

### Requirement: LLM lane Admin API MUST NOT 计入 App usage 统计

`/voice/admin/api/llm-lanes` 与扩展后的 `/ucg/admin/api/ai-config` 为运维型 Admin API，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: Admin 保存不计入 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增该路径计数

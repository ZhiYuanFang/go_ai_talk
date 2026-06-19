## ADDED Requirements

### Requirement: Admin PUT 失效 lane 缓存 MUST NOT 导致进程崩溃

Admin 成功 PUT `/ucg/admin/api/ai-config` 或 `/voice/admin/api/llm-lanes` 后，系统 MUST 失效进程内 lane/profile 缓存以使下一笔 LLM 调用读取新配置；该失效路径 MUST NOT 因 `InvalidateLaneCache` 与 `ProfileStore.InvalidateCache` 互相调用而导致 stack overflow 或进程退出。

#### Scenario: ucg PUT ai-config 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/ucg/admin/api/ai-config` 且持久化成功
- **THEN** ucg-service 进程 MUST 保持运行且后续 GET 同一接口 MUST 返回更新后的配置

#### Scenario: voice PUT llm-lanes 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/voice/admin/api/llm-lanes` 且持久化成功
- **THEN** voice-service 进程 MUST 保持运行且后续 GET MUST 返回更新后的 lane 配置

## MODIFIED Requirements

### Requirement: ucg-service SHALL 扩展 AI 配置 Admin 以包含 polish lane 闸门

`GET/PUT /ucg/admin/api/ai-config` MUST 扩展支持 `provider`、`maxInFlight`、`maxWaiters`（与现有 `visionModel` / `maxImagesPerRequest` 并存或语义对齐为 polish lane 的 `model`）。PUT 成功后 MUST 失效 AI/lane 缓存（`InvalidateAIConfigCache`、ProfileStore 本地 cache 与 `InvalidateLaneCache`，且 MUST 满足「Admin PUT 失效 lane 缓存 MUST NOT 导致进程崩溃」）。

#### Scenario: 管理员更新润笔模型与缓冲池

- **WHEN** 运维 PUT `provider=zhipu`、`visionModel=glm-4.6v-flash`、`maxInFlight=1`、`maxWaiters=15`
- **THEN** ucg-service MUST 持久化且下一笔润笔 MUST 使用新 profile

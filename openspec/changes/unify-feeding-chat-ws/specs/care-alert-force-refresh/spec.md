## ADDED Requirements

### Requirement: care-alert daily 支持 force 强刷

`GET /device/api/care-alert/daily` MUST 支持强制刷新参数（query 名实现为 `force=1` 或等价布尔约定）。当 force 为真时，在通过既有鉴权（App Bearer，网关注入 `X-Internal-Wx-Id>0`）之后，服务端 MUST 删除该 `deviceNo` 当日（Asia/Shanghai）care-alert 日缓存键（经 `cachekit` 既有 `CareAlertDailyKey` builder），再执行与未命中缓存相同的生成/single-flight 流程并写回缓存。无 force 或 force 为假时，行为 MUST 与现网一致（命中日缓存则直接返回且不重跑 LLM）。

#### Scenario: 无 force 命中缓存不重生

- **WHEN** 当日缓存已存在且请求未带 force
- **THEN** 响应 MUST 返回缓存 items，MUST NOT 再次调用 Python 生成

#### Scenario: force 清缓存后重生

- **WHEN** 当日缓存已存在且请求带 force=1（或等价真值），且 wxId>0
- **THEN** 服务端 MUST 删除当日日缓存后重新生成并返回新 items，MUST 写回日缓存

#### Scenario: force 仍要登录

- **WHEN** 请求带 force 但缺少有效 App 鉴权导致 wxId 无效
- **THEN** 服务端 MUST 拒绝请求，MUST NOT 因 force 而跳过鉴权

### Requirement: force 不引入鉴权旁路

care-alert 强刷 MUST NOT 依赖 Admin JWT、MUST NOT 新增仅 Admin 可调的「伪造 wx」接口作为唯一强刷路径。运维调试台 MUST 使用与 App 相同的 daily+force 契约。

#### Scenario: Admin token 不能代替 App 强刷

- **WHEN** 仅携带 Admin JWT 调用 care-alert daily（含 force）
- **THEN** 网关/服务 MUST 按 App 用户接口规则拒绝（与现网 Admin 不可访问 App 接口一致）

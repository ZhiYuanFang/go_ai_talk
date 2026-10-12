## ADDED Requirements

### Requirement: voice Admin 代理 R·predict 飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/predict-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/predict-vectors*`。MUST NOT 直连 Chroma 或新建预测飞轮表。MUST NOT 提供 dedupe 代理。批量 payload MUST 透传上游（含 `predict_channel` 与 `time_answer`）。

#### Scenario: 列出预测飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求预测飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/predict-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库

### Requirement: Hub R·predict 飞轮整理页

Hub SHALL 提供 R·predict 飞轮静态管理页，调用 `/voice/admin/api/predict-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_predict_cues`、document 为整段原文、payload 为 `predict_channel`（`active`|`next`）与 `time_answer`（`clock`|`remain`）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入 R·predict 飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除，并展示两维枚举提示

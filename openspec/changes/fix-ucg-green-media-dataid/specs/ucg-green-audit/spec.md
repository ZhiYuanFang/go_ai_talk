## ADDED Requirements

### Requirement: Green image and video moderation dataId SHALL comply with Alibaba constraints

当 `ucg.green.enabled` 为 true 且系统调用 `GreenModerator.ModerateImageURL` 或 `ModerateVideoURL` 时，若向阿里云 Green 传入 `ServiceParameters.dataId`，该值 MUST 满足：长度不超过 64 字符；字符集仅含大小写英文字母、数字、下划线（`_`）、短划线（`-`）、英文句号（`.`）。系统 MUST NOT 将完整 HTTP(S) URL（含 scheme、`:`、`/` 等）作为 `dataId`。

系统 SHOULD 从媒体 CDN URL 的 path 部分推导合规 `dataId`（例如将 objectKey 中的 `/` 规范为 `_`）。若无法推导合规值，MUST 省略 `dataId` 字段而非传入非法值。

当 Green 返回 `body.Code != 200` 或 HTTP 非 200 时，解析层 MUST 在 error 中包含 business `Code`（及 `Msg` 若 API 提供），以便运维区分参数错误与额度/限流等故障。

#### Scenario: 帖子图片审核使用合规 dataId

- **WHEN** 用户提交带 `social/` 前缀 objectKey 的图片帖且 Phase1 调用 `ModerateImageURL`
- **THEN** Green `ImageModeration` 请求的 `ServiceParameters` MUST NOT 含完整 CDN URL 作为 `dataId`；若含 `dataId` 则 MUST 为规范化 object path（如 `social_2026_06_xxx.jpg`）且长度 ≤64

#### Scenario: 资料头像与私信媒体同步约束

- **WHEN** 资料审核 job 调用头像 `ModerateImageURL`，或私信调用 `ModerateImageURL` / `ModerateVideoURL`
- **THEN** `dataId` 约束 MUST 与帖子媒体相同，MUST NOT 使用完整 URL

#### Scenario: Green 参数错误可观测

- **WHEN** Green 图片 API 因参数校验返回 `body.Code != 200`
- **THEN** ucg-service 日志中的 error MUST 包含 `green image: code <n>` 及 Msg（若存在），MUST NOT 仅返回无 code 的泛化文案

#### Scenario: 合规 dataId 下图片 Phase1 可完成

- **WHEN** 新发图片帖且 Green 配置有效、CDN URL 公网可访问
- **THEN** Phase1 MUST 成功发起 `baselineCheck` 并持久化 `moderation_verdict`（pass 或 reject），MUST NOT 因非法 `dataId` 直接进入 `moderation_failed`（status=5）

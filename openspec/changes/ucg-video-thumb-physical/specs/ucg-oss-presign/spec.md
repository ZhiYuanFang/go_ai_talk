## ADDED Requirements

### Requirement: Video register and server upload SHALL ensure physical first-frame thumb exists

当 `RegisterMedia` 成功登记**新**视频 blob（原视频已在 OSS，`mediaKind=2`）或 `putOSSObject` 成功上传视频（`mediaKind=2`）后，ucg-service MUST 同步调用 `EnsureVideoThumb` 生成物理首帧缩略图对象（`{stem}_thumb.jpg`）。`EnsureVideoThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原视频而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureVideoThumb` 幂等调用 MUST 可接受。

#### Scenario: 新视频 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成 mp4 PUT 且 `RegisterMedia` 成功登记新视频 blob
- **THEN** OSS MUST 同时存在原视频 objectKey 与对应 `{stem}_thumb.jpg`

#### Scenario: 服务端直传视频后生成 thumb

- **WHEN** 服务端成功上传 `mediaKind=2` 视频文件
- **THEN** OSS MUST 存在对应 `_thumb.jpg` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原视频存在但 `EnsureVideoThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

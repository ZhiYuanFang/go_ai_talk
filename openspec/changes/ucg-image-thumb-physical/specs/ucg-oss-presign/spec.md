## ADDED Requirements

### Requirement: Image register and server upload SHALL ensure physical thumb exists

当 `RegisterMedia` 成功登记**新**图片 blob（原图已在 OSS）或 `putOSSObject` 成功上传图片（`mediaKind=1`）后，ucg-service MUST 同步调用 `EnsureImageThumb` 生成物理缩略图对象。`EnsureImageThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原图而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureImageThumb` 幂等调用 MUST 可接受。

#### Scenario: 新图 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成原图 PUT 且 `RegisterMedia` 成功登记新图片 blob
- **THEN** OSS MUST 同时存在原图 objectKey 与对应 `ThumbObjectKey`

#### Scenario: 服务端直传图片后生成 thumb

- **WHEN** `UploadMediaObject` 成功上传 `mediaKind=1` 文件
- **THEN** OSS MUST 存在对应 `_thumb.{ext}` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原图存在但 `EnsureImageThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

## REMOVED Requirements

### Requirement: NormalizeVideo SHALL produce v2 canonical mp4

**Reason**: 服务端同步 ffmpeg 转码占用 ucg 进程资源，且与「客户端交付合规片」策略冲突；转码能力整体删除，后续若需转码须另开独立负载方案。

**Migration**: 客户端与 sim/运维 MUST 上传已通过 v1（或 register 所需版本）验真的 mp4；不得依赖服务端 NormalizeVideo。

### Requirement: Internal upload-video SHALL transcode before OSS

**Reason**: 与全局删除服务端转码一致；internal 路径不再以转码为前置。

**Migration**: 见下方 ADDED「Internal upload-video SHALL validate then PUT」；调用方按响应 `transformVersion` register。

### Requirement: ucg-service container SHALL include ffmpeg

**Reason**: 上传路径不再要求进程内转码；验真以 ffprobe 为准（见 ADDED「container SHALL include ffprobe」）。

**Migration**: 镜像可继续打包 ffmpeg 但不作为上传契约；部署检查改为确认 ffprobe 可用。

## ADDED Requirements

### Requirement: Internal upload-video SHALL validate then PUT without transcode

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload-video`，接受 multipart 视频文件，鉴权 MUST 与现有 `POST /ucg/internal/api/media/upload` 一致（内部网关密钥）。

处理流程 MUST 为：读取 body（上限与 `MaxMediaUploadBytes` 一致）→ 对字节执行 **v1** 验真 → 通过则 PUT OSS（`video/mp4`）→ 响应 `objectKey`、`cdnUrl`、`contentHash`（SHA-256 hex 小写，对 **OSS 上最终字节** 即原始合规字节计算）、`transformVersion=v1`。

v1 验真失败 MUST 返回 4xx 且 MUST NOT 在 OSS 创建对象。MUST NOT 调用 `NormalizeVideo` 或任何进程内转码。本接口 MUST NOT 自动 `RegisterMedia`。

#### Scenario: Internal upload returns v1 object

- **WHEN** 内部密钥有效且上传通过 v1 验真的 mp4
- **THEN** 响应 MUST 含 objectKey、contentHash 与 `transformVersion=v1`，且 OSS 对象字节 MUST 与上传字节一致

#### Scenario: Internal upload unauthorized

- **WHEN** 未提供有效内部密钥
- **THEN** MUST 返回 403

#### Scenario: Internal upload rejects non-compliant video

- **WHEN** 上传未通过 v1 验真（含不可解码）
- **THEN** MUST 返回 4xx 且 MUST NOT 留下 OSS 对象，MUST NOT 转码

### Requirement: ucg-service container SHALL include ffprobe for validation

`manifest/docker/Dockerfile.ucg-service` 构建的镜像 MUST 包含可执行的 `ffprobe`，供视频验真使用。部署 ucg-service MUST 使用含 ffprobe 的镜像。镜像 MAY 仍包含 `ffmpeg` 二进制，但上传路径 MUST NOT 依赖其做转码。

#### Scenario: Probe available in container

- **WHEN** ucg-service 容器启动
- **THEN** 进程 MUST 能成功执行 `ffprobe -version`

## ADDED Requirements

### Requirement: NormalizeVideo SHALL produce v2 canonical mp4

ucg-service MUST 提供 `NormalizeVideo`（或等价导出函数），将任意可解码输入转码为 **v2 canonical** mp4：

- 视频：libx264，profile **main**，pix_fmt **yuv420p**
- 音频：**aac**；若输入无音轨 MUST 补 **静音 AAC** 轨并与视频时长对齐
- 容器：mp4，**movflags +faststart**
- 输出 MUST 通过 `v2` 验真规则

转码 MUST 使用进程内 `ffmpeg`/`ffprobe`（非 OSS 侧处理）。失败 MUST 返回错误且 MUST NOT 上传半成品至 OSS。

#### Scenario: Transcode silent video adds AAC

- **WHEN** 输入文件无音轨且调用 `NormalizeVideo`
- **THEN** 输出 MUST 含 AAC 音轨且 MUST 通过 v2 验真

#### Scenario: Transcode output is faststart mp4

- **WHEN** 输入为 moov 后置的 h264+AAC mp4
- **THEN** 输出 MUST 满足 v2 faststart 要求

### Requirement: Internal upload-video SHALL transcode before OSS

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload-video`，接受 multipart 视频文件，鉴权 MUST 与现有 `POST /ucg/internal/api/media/upload` 一致（内部网关密钥）。

处理流程 MUST 为：读取 body（上限与 `MaxMediaUploadBytes` 一致）→ `NormalizeVideo` → PUT OSS（`video/mp4`）→ 响应 `objectKey`、`cdnUrl`、`contentHash`（SHA-256 hex 小写，对 **OSS 上最终字节** 计算）。

本接口 MUST NOT 自动 `RegisterMedia`；调用方 MUST 自行 register（含 `transformVersion=v2`）。

#### Scenario: Internal upload returns canonical object

- **WHEN** 内部密钥有效且上传可解码视频
- **THEN** 响应 MUST 含 objectKey 与 contentHash，且 OSS 对象 MUST 通过 v2 验真

#### Scenario: Internal upload unauthorized

- **WHEN** 未提供有效内部密钥
- **THEN** MUST 返回 403

#### Scenario: Internal upload transcode failure

- **WHEN** ffmpeg 无法解码输入
- **THEN** MUST 返回 5xx 或 4xx 明确错误且 MUST NOT 留下 OSS 对象

### Requirement: ucg-service container SHALL include ffmpeg

`manifest/docker/Dockerfile.ucg-service` 构建的镜像 MUST 包含可执行的 `ffmpeg` 与 `ffprobe`，供验真与转码使用。部署 ucg-service  MUST 使用含 ffmpeg 的镜像方可启用本能力。

#### Scenario: Transcode available in container

- **WHEN** ucg-service 容器启动且配置启用视频转码
- **THEN** 进程 MUST 能成功执行 `ffmpeg -version` 与 `ffprobe -version`

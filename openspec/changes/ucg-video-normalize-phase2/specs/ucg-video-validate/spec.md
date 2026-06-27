## ADDED Requirements

### Requirement: ProbeVideoDecodable SHALL distinguish transcodable from corrupt uploads

ucg-service MUST 提供对上传字节的 **可解码探测**（如 `ProbeVideoDecodable`），用于 Web 视频代理上传在 **v1 验真失败** 后判断是否进入服务端转码：

- ffprobe MUST 能解析容器与 streams
- streams MUST 含至少一条 `codec_type=video` 的视频轨

不满足上述条件 MUST 视为 **不可解码**，Web upload MUST 返回 4xx 且 MUST NOT 调用 `NormalizeVideo`。

本探测 MUST NOT 替代 v1/v2 验真；仅用于 B 分支（转码兜底）门禁。

#### Scenario: HEVC mp4 is decodable for transcode fallback

- **WHEN** 上传 mp4 含 h265/hevc 视频轨，ffprobe 成功解析
- **THEN** `ProbeVideoDecodable` MUST 成功（即使 v1 验真因非 h264 失败）

#### Scenario: Corrupt file is not decodable

- **WHEN** 上传字节无法被 ffprobe 解析
- **THEN** `ProbeVideoDecodable` MUST 失败且 Web upload MUST NOT PUT OSS

## MODIFIED Requirements

### Requirement: Web video proxy upload SHALL validate v1 before OSS PUT

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 时 MUST 对上传字节执行 **v1 验真优先** 的分支处理：

1. **v1 通过**：MUST 将 **原始上传字节** 直传 OSS（与 Phase 1 一致）；`contentHash` MUST 为原始字节 SHA-256；响应 MUST 提示 `transformVersion=v1`。
2. **v1 未通过且可解码**：MUST 调用 `NormalizeVideo` 转码为 v2 canonical mp4 后 PUT OSS；`contentHash` MUST 为 **转码后 OSS 字节** SHA-256；响应 MUST 提示 `transformVersion=v2`；register MUST 使用 `v2`（非 `v1`）。
3. **v1 未通过且不可解码**：MUST 返回 4xx 且 MUST NOT 在 OSS 上创建对象。

v1 已合规但缺少 faststart 的对象 MUST 走路径 1（直传 `v1`），MUST NOT 仅为 faststart 触发 remux 或转码。

#### Scenario: Web compliant video uploads direct v1

- **WHEN** Web 上传满足 v1 规则的 mp4（可无 faststart）
- **THEN** 响应 MUST 含 `objectKey`、`cdnUrl`、`contentHash`，OSS MUST 存在与上传字节一致的对象，且 MUST 含 `transformVersion=v1`

#### Scenario: Web non-compliant but decodable transcodes to v2

- **WHEN** Web 上传 HEVC mp4 或无 AAC 音轨但 ffprobe 可解析且含视频轨
- **THEN** API MUST 成功返回且 OSS 对象 MUST 通过 v2 验真，响应 MUST 含 `transformVersion=v2` 与转码后 `contentHash`

#### Scenario: Web undecodable video rejected

- **WHEN** Web 上传损坏或非视频文件
- **THEN** API MUST 返回 4xx 且 OSS MUST NOT 存在新对象

#### Scenario: v1 compliant without faststart stays direct

- **WHEN** Web 上传 h264 Main/Baseline + AAC 的 mp4 但 moov 不在 mdat 前（v1 允许）
- **THEN** MUST 直传原始字节且 `transformVersion` MUST 为 `v1`（MUST NOT 转码）

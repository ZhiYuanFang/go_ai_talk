## MODIFIED Requirements

### Requirement: Web video proxy upload SHALL validate v1 before OSS PUT

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 时 MUST 对上传字节执行 **v1 验真**，且 MUST NOT 进行服务端转码：

1. **v1 通过**：MUST 将 **原始上传字节** 直传 OSS；`contentHash` MUST 为原始字节 SHA-256；响应 MUST 提示 `transformVersion=v1`。
2. **v1 未通过**（含不可解码、编码不合规等）：MUST 返回 4xx 且 MUST NOT 在 OSS 上创建对象。

MUST NOT 在 v1 失败后调用 `NormalizeVideo`、`ProbeVideoDecodable` 转码门禁或任何进程内转码后再 PUT。v1 已合规但缺少 faststart 的对象 MUST 直传 `v1`（faststart 非 v1 强制项）。

#### Scenario: Web compliant video uploads direct v1

- **WHEN** Web 上传满足 v1 规则的 mp4（可无 faststart）
- **THEN** 响应 MUST 含 `objectKey`、`cdnUrl`、`contentHash`，OSS MUST 存在与上传字节一致的对象，且 MUST 含 `transformVersion=v1`

#### Scenario: Web non-compliant video rejected without transcode

- **WHEN** Web 上传 HEVC mp4 或无 AAC 音轨或其它未通过 v1 验真的文件（无论 ffprobe 是否可解析）
- **THEN** API MUST 返回 4xx，MUST NOT 创建 OSS 对象，MUST NOT 返回 `transformVersion=v2` 转码成功响应

#### Scenario: Web undecodable video rejected

- **WHEN** Web 上传损坏或非视频文件
- **THEN** API MUST 返回 4xx 且 OSS MUST NOT 存在新对象

#### Scenario: v1 compliant without faststart stays direct

- **WHEN** Web 上传 h264 Main/Baseline + AAC 的 mp4 但 moov 不在 mdat 前（v1 允许）
- **THEN** MUST 直传原始字节且 `transformVersion` MUST 为 `v1`（MUST NOT 转码）

## REMOVED Requirements

### Requirement: ProbeVideoDecodable SHALL distinguish transcodable from corrupt uploads

**Reason**: Web 上传不再提供「可解码则转码」B 分支；v1 失败一律拒绝，无需可解码探测门禁。

**Migration**: 删除或停用对上传路径的 `ProbeVideoDecodable` 调用；客户端须提交 v1 合规片源。

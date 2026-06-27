## ADDED Requirements

### Requirement: ucg-service SHALL validate video by transformVersion with forked rules

ucg-service MUST 提供按 `transformVersion` 分叉的视频 ffprobe 验真能力。`v1` 与 `v2` 规则 **部分对齐、刻意分叉**：通过 `v1` 验真 **不** 等价于 canonical 合规。

**v2（canonical）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）
- **faststart**：moov MUST 位于 mdat 之前（可播放渐进下载）
- 大小 ≤ ucg-service 配置的单文件上传上限

**v1（Web Phase 1 宽规）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main 或 Baseline**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）；**无音轨 MUST 拒绝**（Phase 1 不补轨）
- faststart **不** 强制
- 大小 ≤ 单文件上传上限

验真 MUST 支持对内存字节与 OSS objectKey（Range 读取 + ffprobe）执行。未列出的 `transformVersion` MUST NOT 套用 v1/v2 规则。

#### Scenario: v2 rejects Baseline profile

- **WHEN** OSS 对象为 h264 Baseline + AAC + faststart 的 mp4 且 register 请求 `transformVersion=v2`
- **THEN** 验真 MUST 失败且 register MUST 返回 4xx

#### Scenario: v1 accepts Baseline with AAC

- **WHEN** Web 上传 h264 Baseline + AAC 的 mp4（可无 faststart）且验真版本为 v1
- **THEN** 验真 MUST 通过

#### Scenario: v1 rejects missing audio track

- **WHEN** 视频仅有视频轨、无音轨，且验真版本为 v1
- **THEN** 验真 MUST 失败且 MUST NOT 直传 OSS（Web upload）或 MUST 拒绝 register

#### Scenario: v2 rejects missing AAC

- **WHEN** 视频含非 AAC 音轨（如 mp3）且验真版本为 v2
- **THEN** 验真 MUST 失败

#### Scenario: v2 rejects missing faststart

- **WHEN** mp4 满足 h264 Main + AAC 但 moov 不在 mdat 前且验真版本为 v2
- **THEN** 验真 MUST 失败

### Requirement: Web video proxy upload SHALL validate v1 before OSS PUT

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 时 MUST 在 **PUT OSS 之前** 对上传字节执行 `v1` 验真。通过则 MUST 将 **原始上传字节** 直传 OSS（Phase 1 不转码）。失败 MUST 返回 4xx 且 MUST NOT 在 OSS 上创建对象。

#### Scenario: Web compliant video uploads

- **WHEN** Web 上传满足 v1 规则的 mp4
- **THEN** 响应 MUST 含 `objectKey` 与 `cdnUrl`，且 OSS MUST 存在与上传字节一致的对象

#### Scenario: Web non-compliant video rejected

- **WHEN** Web 上传 HEVC 或无音轨 mp4
- **THEN** API MUST 返回 4xx 且 OSS MUST NOT 存在新对象

### Requirement: Video RegisterMedia SHALL validate OSS object by transformVersion

`RegisterMedia` 在 `mediaKind=2` 时 MUST：

- 仅接受 `transformVersion` 为 `v1` 或 `v2`；其他值（含 `sim-raw`）MUST 返回 400
- 在登记 blob/ownership 之前 MUST 对 `objectKey` 对应 OSS 对象执行与请求 `transformVersion` 匹配的验真
- 验真失败 MUST 返回 4xx 且 MUST NOT 完成 blob 登记

`mediaKind=1` 行为 MUST 不变（图片 thumb 逻辑不受本 requirement 影响）。

#### Scenario: Native v2 register succeeds

- **WHEN** 客户端 PUT 满足 v2 规则的 mp4 且 `RegisterMedia` 带 `transformVersion=v2` 与正确 contentHash
- **THEN** 登记 MUST 成功

#### Scenario: sim-raw register rejected

- **WHEN** `RegisterMedia` 请求 `transformVersion=sim-raw` 且 `mediaKind=2`
- **THEN** MUST 返回 400 且 MUST NOT 登记

#### Scenario: v1 register validates relaxed rules

- **WHEN** OSS 对象为 v1 合规（Baseline + AAC、无 faststart）且 register `transformVersion=v1`
- **THEN** 登记 MUST 成功

#### Scenario: Register v2 on v1-only object fails

- **WHEN** OSS 对象仅满足 v1（如无 faststart）但 register `transformVersion=v2`
- **THEN** MUST 返回 4xx

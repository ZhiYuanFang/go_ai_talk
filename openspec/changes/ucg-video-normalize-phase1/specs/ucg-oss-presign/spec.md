## ADDED Requirements

### Requirement: Video media SHALL use transformVersion v1 or v2 with validation gate

视频 blob 索引 MUST 使用 `transformVersion` 区分管线产出：

- `v2`：canonical（H.264 Main + AAC + faststart）；原生客户端转码与 sim 服务端转码产出 MUST 使用 `v2` register
- `v1`：Web Phase 1 宽验真直传；**非 canonical**（可 Baseline、可无 faststart），但 **必须有 AAC 音轨**

`contentHash` MUST 为 OSS 上最终对象字节的 SHA-256 hex 小写。`v1` 与 `v2` MUST NOT 跨版本 dedup 命中（键为 `(contentHash, transformVersion)`）。

新写入 MUST NOT 使用 `sim-raw` 或其他未登记版本。

#### Scenario: v1 and v2 blobs do not dedup each other

- **WHEN** 同一源文件分别作为 v1 直传与 v2 转码后 contentHash 不同或 version 不同
- **THEN** `media/resolve` MUST NOT 跨 v1/v2 返回 hit

### Requirement: Web video upload response SHALL include contentHash for register

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 且上传成功时，响应 `data` MUST 额外含 `contentHash`（64 位小写 hex，对 OSS 对象字节计算，Phase 1 与上传字节一致），以便 Web 客户端 `RegisterMedia` 使用 `transformVersion=v1`。

#### Scenario: Web upload returns hash

- **WHEN** Web 成功上传 v1 合规视频
- **THEN** JSON `data` MUST 含 `contentHash` 且长度为 64

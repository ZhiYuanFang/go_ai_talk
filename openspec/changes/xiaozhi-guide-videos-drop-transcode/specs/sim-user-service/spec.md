## MODIFIED Requirements

### Requirement: sim T4 video upload SHALL use ucg internal transcode API

sim-user-service 在上传 T4 视频至 UCG 时 MUST 经 HTTP 调用 ucg-service `POST /ucg/internal/api/media/upload-video`（与 device internal 鉴权方式一致），MUST 使用返回的 `objectKey`、`contentHash` 与 `transformVersion` 调用 App API `RegisterMedia`（`mediaKind=2`、`dedupHit=false`）。首版服务端不再转码时，响应 `transformVersion` MUST 为 `v1`，register MUST 使用同一版本。

MUST NOT 再使用 `transformVersion=sim-raw` 或 presign 直传未验真字节。MUST NOT 假设 internal 接口会将任意片源转码为 v2；T4 源片 MUST 已满足 ucg v1 验真规则。

#### Scenario: T4 registers after internal validate-and-put

- **WHEN** internal 上传成功且响应含 `transformVersion=v1`
- **THEN** 后续 register MUST 使用 `transformVersion=v1` 且 MUST 通过 ucg 侧 v1 验真

#### Scenario: Internal upload failure skips job

- **WHEN** internal 上传因验真失败或其它错误失败
- **THEN** T4 流水线 MUST 将 job 标为 skipped 且 MUST NOT 使用 presign 回退直传

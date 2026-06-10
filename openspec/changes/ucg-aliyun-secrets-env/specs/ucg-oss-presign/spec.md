## MODIFIED Requirements

### Requirement: OSS presign SHALL use pang-bao bucket with social/ prefix

ucg-service SHALL provide presigned upload for bucket `pang-bao`, region `cn-beijing`, endpoint `oss-cn-beijing.aliyuncs.com`, generating objectKey with prefix `social/`. Database and API DTOs MUST store objectKey only; CDN display URL is `https://resorce.cuplay.top/{objectKey}`.

OSS AccessKey credentials MUST NOT be stored as plaintext in `manifest/config/config.ucg-service.yaml`. The yaml fields `ucg.oss.accessKeyId` and `ucg.oss.accessKeySecret` MUST be empty in the repository. At runtime, credentials MUST be supplied via environment variables `UCG_OSS_ACCESS_KEY_ID` and `UCG_OSS_ACCESS_KEY_SECRET` (typically injected through `manifest/docker/.env.*` and Docker Compose).

#### Scenario: 获取 presign

- **WHEN** 客户端 `POST /ucg/app/api/media/presign` with media kind and extension，且 ucg-service 已通过 env 配置有效 OSS 凭证
- **THEN** 响应 SHALL 含 uploadUrl、objectKey（以 `social/` 开头），且 SHALL NOT 要求客户端自定义 bucket

#### Scenario: DB 仅存 objectKey

- **WHEN** 帖子媒体写入 `ucg_post_media`
- **THEN** 行 SHALL 仅保存 objectKey 字段，且 SHALL NOT 保存完整 CDN URL

#### Scenario: 未配置 OSS 凭证时 presign 失败

- **WHEN** 容器 env 与 yaml 均未提供有效 `UCG_OSS_ACCESS_KEY_*`
- **THEN** presign 接口 SHALL 返回明确错误且 SHALL NOT 使用硬编码或 yaml 明文 fallback

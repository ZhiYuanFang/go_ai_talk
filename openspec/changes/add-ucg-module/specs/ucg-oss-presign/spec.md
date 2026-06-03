## ADDED Requirements

### Requirement: OSS presign SHALL use pang-bao bucket with social/ prefix

ucg-service SHALL provide presigned upload for bucket `pang-bao`, region `cn-beijing`, endpoint `oss-cn-beijing.aliyuncs.com`, generating objectKey with prefix `social/`. Database and API DTOs MUST store objectKey only; CDN display URL is `https://resorce.cuplay.top/{objectKey}`.

Credentials SHALL be configured in `manifest/config/config.ucg-service.yaml` (overridable by env):

| 项 | 值 |
|----|-----|
| AccessKey ID | `LTAI5t6tomJZp4im2H32FSMT` |
| AccessKey Secret | `LVCECT4exrGkkhI85HmyD4P2e6wJZW` |

#### Scenario: 获取 presign
- **WHEN** 客户端 `POST /ucg/app/api/media/presign` with media kind and extension
- **THEN** 响应 SHALL 含 uploadUrl、objectKey（以 `social/` 开头），且 SHALL NOT 要求客户端自定义 bucket

#### Scenario: DB 仅存 objectKey
- **WHEN** 帖子媒体写入 `ucg_post_media`
- **THEN** 行 SHALL 仅保存 objectKey 字段，且 SHALL NOT 保存完整 CDN URL

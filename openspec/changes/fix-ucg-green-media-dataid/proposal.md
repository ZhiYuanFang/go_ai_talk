## Why

UCG 机审调用阿里云 Green 图片/视频 API 时，将完整 CDN URL 作为 `dataId` 传入，违反阿里云约束（长度 ≤64、字符集仅 `A-Za-z0-9_.-`），导致 `ImageModeration` / `baselineCheck` 参数校验失败。表现为：文本审核正常、视频可能正常，但**图片帖 Phase1 失败并落入 status=5（moderation_failed）**，控制台无成功 baselineCheck 记录。

## What Changes

- 在 `green_client.go` 为图片/视频审核生成**合规 `dataId`**：从媒体 URL 提取 object path 并规范化（`/` → `_`），超长时截断或省略；**禁止**再使用完整 URL 作为 `dataId`。
- 同步改进 `parseImageModeration` / `parseVideoModeration` 错误信息，输出 `body.Code` 与 `Msg`，便于运维定位参数/API 失败。
- **不修改** `GreenModerator` 对外接口签名（call site 零改）。
- **不修改** green-once / moderation_failed 状态机；已 status=5 的历史帖需人工或重新提交处理。
- **不包含** `infoType`、OSS 授权检测、`postImageCheck` 服务切换（若 dataId 修复后仍失败，另开变更）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-green-audit`：明确 `ModerateImageURL` / `ModerateVideoURL` 调用 Green 时，若传入 `dataId` MUST 符合阿里云长度与字符集约束；MUST NOT 使用完整 HTTP(S) URL 作为 `dataId`。

## Impact

- **代码**：`internal/services/ucg/green_client.go`（主改动）；`audit_moderation.go` 等 call site **无需改动**。
- **服务**：仅 `ucg-service` 部署生效。
- **API/DB**：无协议或表结构变更；修复后新发带图帖/头像/私信图应能正常进入 Green 并完成 Phase1。
- **运维**：部署后 status=5 存量帖不会自动重审；runbook 可补充验证步骤。

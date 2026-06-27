## Why

`GET /ucg/app/api/profile/{wxId}` 公开 profile 读路径未填充 `followingCount` / `followerCount` / `postCount`，客户端他人主页始终显示「关注 0」。`profile/me` 经 `mergeProfileForAuthor` → `enrichProfileStats` 已有真实统计，两路径不一致。

## What Changes

- `GetPublicProfile` 返回前 MUST 调用已有 `enrichProfileStats`，与 `profile/me` 统计语义对齐。
- 响应字段名不变（`UcgProfileRes` 已有 `omitempty` 字段），属补全已有契约，非新接口版本。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-app-profile`：公开 profile 读路径返回社交统计。

## Impact

- `internal/services/ucg/profile.go`（`GetPublicProfile`）
- 无 DB 迁移、无 Redis、无 gateway 变更
- Flutter 他人主页无需改代码，后端补字段后自动生效

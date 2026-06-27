## MODIFIED Requirements

### Requirement: 公开 profile 返回社交统计

`GET /ucg/app/api/profile/{wxId}` MUST 在响应中返回与 `GET /ucg/app/api/profile/me` 一致的 `followingCount`、`followerCount`、`postCount`，数值 MUST 经 ucg 库内实时 COUNT（`ucg_follow`、`ucg_post`），与 `enrichProfileStats` 语义相同。

#### Scenario: 他人主页展示关注数

- **WHEN** 客户端请求 `GET /ucg/app/api/profile/{wxId}` 且该 wx 存在 profile 行
- **THEN** 响应 MUST 包含 `followingCount` 等于 `COUNT(ucg_follow WHERE follower_wx_id = wxId)`
- **AND** 响应 MUST 包含 `followerCount` 等于 `COUNT(ucg_follow WHERE followee_wx_id = wxId)`
- **AND** 响应 MUST 包含 `postCount` 等于 `COUNT(ucg_post WHERE author_wx_id = wxId)`

#### Scenario: profile 不存在

- **WHEN** 请求的 wxId 无 `ucg_profile` 行
- **THEN** MUST 返回 404，行为与变更前一致

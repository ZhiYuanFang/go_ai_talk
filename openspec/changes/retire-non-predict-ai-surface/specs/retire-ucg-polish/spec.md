## ADDED Requirements

### Requirement: UCG 润笔 polish MUST 删除

ucg-service MUST 删除 `POST /ucg/app/api/posts/polish` 及其控制器/服务实现（`PolishPostText` 等）、polish lane 配置与 Admin AI config 中仅服务润笔的项。App 调用原 polish 路径 MUST 不再执行润笔 LLM。

#### Scenario: polish 接口不可用

- **WHEN** 客户端 POST 原 `/ucg/app/api/posts/polish`
- **THEN** 服务端 MUST NOT 再执行润笔生成逻辑

### Requirement: UCG polish AI 额度 MUST 删除

ucg-service MUST 删除 polish 月度额度（`ai_quota_*` 在 UCG 库中的 polish 列与 App/Admin 额度 API）。MUST NOT 删除 UCG 社区本体能力。

#### Scenario: UCG ai-quota API 下线

- **WHEN** 客户端请求原 `GET /ucg/app/api/ai-quota`
- **THEN** 服务端 MUST NOT 再返回 polish 额度业务响应

### Requirement: UCG 社区本体 MUST 保留

删除润笔与额度时，系统 MUST 保留媒体、发帖/改删、Feed、关注/点赞/投票/评论、私信与 WS chat、资料、Green 审核与 Admin 审帖等社区能力。

#### Scenario: 发帖与 Feed 仍可用

- **WHEN** 用户创建帖子或拉取 recommend/following Feed
- **THEN** 接口 MUST 仍按既有社区契约工作，MUST NOT 因删除 polish 而 5xx

## MODIFIED Requirements

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」与「批量驳回」按钮；批量驳回前 MUST 经用户确认。`status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图（非仅第一条）；图片 SHALL 支持点击后在 modal 中查看原图（`cdnUrl`）；视频 SHALL 展示首帧缩略图并支持点击后在 modal 内播放（`<video controls>` + `cdnUrl`）。

同一页面 MUST 提供与「动态审查」并列的 **「资料机审失败」** Tab（或与 AI 配置/动态审查同级 Tab 导航），用于展示 `ProfileJobStatusModerationFailed` 的 profile audit job 列表及人工操作入口（详见 `ucg-admin-profile-moderation` 规格）。

#### Scenario: 动态审查批量驳回

- **WHEN** 管理员在「动态审查」勾选多条未驳回帖子并确认批量驳回
- **THEN** 系统 SHALL 调用批量驳回 API 且刷新列表

#### Scenario: 管理页含资料机审 Tab

- **WHEN** 管理员打开 ucg-admin.html 并已通过口令登录
- **THEN** 页面 SHALL 展示「资料机审失败」Tab 入口，且 SHALL NOT 与「动态审查」互斥隐藏

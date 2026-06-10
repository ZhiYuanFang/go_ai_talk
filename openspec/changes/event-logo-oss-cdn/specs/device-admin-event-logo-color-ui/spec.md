## MODIFIED Requirements

### Requirement: 管理页 Logo 预览使用 CDN URL

管理页用于 `<img src>` 的 logo 地址 SHALL 直接使用 API 返回的 CDN 绝对 URL；SHALL NOT 再拼接当前页面 origin 与 `/ai_talk_images/` path。

#### Scenario: API 返回 CDN logo 时展示

- **WHEN** `GET /device/admin/api/event/list` 返回 `logo` 为 `https://resorce.cuplay.top/event/...`
- **THEN** 页面 SHALL 将该 URL 直接用于 `<img src>`

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

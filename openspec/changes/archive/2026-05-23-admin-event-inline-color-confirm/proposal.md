## Why

事件管理页（`admin.html`）行内修改色调时，使用隐藏的原生 `<input type="color">` 并在 **`change` 事件立即提交**，用户无法在确认前预览或取消，体验差且易误触保存。需要在色调选择流程中提供明确的 **「确定」**（及 **「取消」**）操作后再写库。

## What Changes

- 在 `resource/public/admin.html` 为行内色调编辑增加小型浮层/面板：展示取色器、当前色值预览、**确定** / **取消** 按钮。
- 仅在用户点击 **确定** 后调用现有 `submitEventRowUpdate`（`POST .../event/update` multipart）；**取消** 关闭面板且不请求接口。
- 保留弹窗「编辑事件」中的色调字段与提交逻辑不变；Logo 行内上传仍为选文件即提交（本变更不调整 Logo 交互）。
- **不修改** 后端 API 与 Redis 缓存刷新逻辑（事件缓存同步问题另立变更处理）。

## Capabilities

### New Capabilities

- `admin-event-inline-color-confirm`：事件管理页行内色调编辑须用户确认后再保存。

### Modified Capabilities

（无。）

## Impact

- **前端**：`resource/public/admin.html`（样式 + 行内色调交互脚本）。
- **后端 / 部署**：无变更。

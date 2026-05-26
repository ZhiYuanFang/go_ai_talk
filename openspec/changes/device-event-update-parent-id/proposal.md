## Why

`device-event-hierarchy` 已在 `event` 表用 `parent_id` 建树，并在**新增**时支持指定父节点；但 **`UpdateEvent` 第一期禁止修改 `parent_id`**，管理端编辑事件时也无法调整层级。运营在事件归类纠错（例如将误挂在根下的节点挪到「换尿布」下）时只能删建，成本高且易丢 logo/别名配置。

## What Changes

- **device-service**：`UpdateEvent` 支持可选修改 **`parent_id`**（`0` = 升为根）；校验父节点存在、禁止自引用与**成环**、移动后仍满足**同父下 `name` 唯一**；成功后照旧触发事件 Redis 缓存全量重建。
- **管理端 `admin.html`**：编辑事件时提供**父事件选择**（含「无父 / 根」）；提交 `POST /device/admin/api/event/update` 时携带 `parentId`；禁止将父选为自身（前端过滤，后端仍强校验）。
- **契约**：`DeviceAdminContract.UpdateEvent` 与 HTTP multipart 字段文档补充 `parentId`；新增业务错误语义（父不存在、成环、有子节点时是否允许移动——见 design）。
- **不迁移** `history` 行；已有 `history.event_id` 不变，仅事件字典树结构变化。

## Capabilities

### New Capabilities

- `device-event-update-parent-id`：事件更新 API 修改 `parent_id` 的规则、成环检测与同父唯一性。

### Modified Capabilities

- `device-event-hierarchy`（change 内增量）：解除「第一期禁止修改 `parent_id`」；`UpdateEvent` 与验收场景对齐。
- `device-admin-event-tree-ui`（change 内增量）：编辑表单增加父事件选择器。
- `device-event-cache-rebuild-on-mutate`：父节点变更后缓存快照中的 `parentId` 与库一致（行为不变，补充场景）。

## Impact

- **数据库**：`ai_voice_device.event.parent_id` 已存在；**无**新列；可选后续加 CHECK/触发器，本期以应用层成环检测为准。
- **device-service**：`internal/services/device/admin.go`（`UpdateEvent`）、`internal/controller/device_admin_event.go`、`api/v1/device_admin_http.go`、`internal/services/contracts/runtime_contracts.go`、`internal/services/device/admin_http_client.go`。
- **前端**：`resource/public/admin.html` 事件编辑弹窗。
- **voice-service**：无逻辑变更；`ListEvents` 缓存刷新后自动获得新树形索引。
- **依赖前置**：`device-event-hierarchy` 已实现项（`parent_id` 读写、树形列表、AddEvent parentId）应已部署或与本变更同批发布。

## Non-goals

- 批量拖拽排序、跨环境事件树同步。
- 修改父节点时自动迁移子节点 logo/color。
- 历史 `history.event_name` 批量重写。

## 1. device-service API 与领域逻辑

- [x] 1.1 `DeviceAdminContract.UpdateEvent` 增加 `parentID int64` 参数；`admin_http_client` 同步
- [x] 1.2 `UpdateEvent`：解析目标 `parent_id`；父存在校验；后代集合成环检测；同父 `name` 查重；写库 `parent_id`
- [x] 1.3 新增 `ErrEventParentInvalid` / `ErrEventParentCycle`（或等价 sentinel）与 `writeDeviceAdminEventErr` 映射
- [x] 1.4 `device_admin_event.go`：`event/update` 解析表单 `parentId` 并传入 `UpdateEvent`
- [x] 1.5 `api/v1/device_admin_http.go`：`DeviceAdminEventUpdateReq` 文档补充 `ParentId` 字段说明

## 2. 管理端 admin.html

- [x] 2.1 编辑弹窗增加父事件下拉（根 + 扁平列表过滤自身与后代）
- [x] 2.2 `eventFormMode === 'edit'` 提交时 `fd.append('parentId', ...)`
- [x] 2.3 成环/父无效错误展示后端 message

## 3. 验证与 OpenSpec

- [ ] 3.1 手动：将叶子从根下挂到「换尿布」；再提回根；尝试选子孙为父应失败
- [ ] 3.2 手动：改父后刷新列表与 voice 侧事件树（缓存含新 parentId）
- [x] 3.3 `openspec validate device-event-update-parent-id --strict` 通过
- [x] 3.4 更新 `openspec/changes/device-event-hierarchy/tasks.md` 1.3 说明（可选：勾选「父 id 可改」或链到本 change）

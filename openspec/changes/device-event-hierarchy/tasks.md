## 1. device-service 数据层与 API

- [x] 1.1 `eventListFields` / `RebuildEventCache` 增加 `parent_id`；`ListEvents` 返回 `parentId`
- [x] 1.2 `AddEvent` 支持 `parentId`：校验父存在、同父 name 唯一、不复制父 logo/color
- [x] 1.3 `UpdateEvent` 查重改为同 `(parent_id, name)`；第一期禁止修改 `parent_id`
- [x] 1.4 `DeleteEvent`：存在 `parent_id=id` 子行时返回业务错误
- [x] 1.5 `device_admin_event.go` / API 类型：`event/add` multipart 增加 `parentId` 表单字段
- [x] 1.6 更新 `DeviceAdminContract` 与 HTTP client 签名（若 AddEvent 签名变更）
- [x] 1.7 业务代码确认不读写 `child_ids` / `ChildIds`

## 2. 管理端 admin.html

- [x] 2.1 实现 `buildEventTree(flatList)` 与树形表格渲染（缩进、递归子行）
- [x] 2.2 每行增加「新增子事件」按钮；弹窗标题/模式区分根与子，提交带 `parentId`
- [x] 2.3 子事件弹窗不预填父 logo/color；保留行内 logo/color 编辑
- [x] 2.4 删除失败时展示后端「有子节点」类错误提示

## 3. voice-service 事件树与匹配

- [x] 3.1 新增 `buildEventTreeIndex(events)`：`childrenByParent`、`depthById`、`hasChildren`
- [x] 3.2 重构 `extractEventFromText`：支持候选集过滤 + 深度/名长排序
- [x] 3.3 新增 `pendingChildEventState` 及 get/set/clear（按 deviceNo）
- [x] 3.4 实现 `resolveEventForRecord`：pending 子树匹配、非叶子追问文案、`finishTalk=false`
- [x] 3.5 接入 `handleActionRecord` / `handleUnifiedIntentAction` / `resolveEventFromUnifiedIntent` 统一走新解析
- [x] 3.6 确认仅叶子 `event_id` 写入 history；追问轮次不写库
- [x] 3.7 轻量更新 DeepSeek 事件列表提示（层级说明；追问仍由代码主导）

## 4. 验证与收尾

- [ ] 4.1 手动验证：管理端创建 换尿布→大便/小便 树；子节点独立 logo/color
- [ ] 4.2 手动验证：voice「换尿布」→ 追问 →「大便」落库；「换尿布+大便」一次落库
- [ ] 4.3 手动验证：删有子父节点失败；删叶子成功且缓存刷新含 parentId
- [x] 4.4 `openspec validate device-event-hierarchy --strict` 通过

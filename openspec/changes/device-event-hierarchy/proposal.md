## Why

当前 `event` 表为扁平字典，无法表达「换尿布 → 大便 / 小便」等层级关系；语音命中父类事件时会直接落库，无法引导用户细分。需要在设备域引入基于 `parent_id` 的通用事件树，并在管理端与 voice 链路中一致支持「非叶子节点追问、叶子节点记录」。

## What Changes

- `event` 表以 **`parent_id` 为唯一层级字段**（`0` 表示根）；**不再读写 `child_ids`**（库中若存在该列则忽略）。
- **device-service**：`AddEvent` 支持 `parentId`；`ListEvents` / Redis 缓存重建含 `parentId`；删除时若存在子节点则拒绝；同父下 `name` 唯一（替代全局唯一）。
- **管理端 `admin.html`**：事件列表改为树形展示；每行支持「新增子事件」；子事件 **不继承** 父节点 `logo` / `color`，各自独立配置。
- **voice-service**：构建事件树索引；匹配时子节点优先；命中**有子节点**的事件时不写 `history`，进入多轮追问（动态拼接直接子节点名称）；仅**叶子**可落库；pending 状态为进程内存态，会话过期后不恢复。
- 不迁移历史 `history` 数据。

## Capabilities

### New Capabilities

- `device-event-hierarchy`: 事件表 `parent_id` 通用树、device API 层级 CRUD 规则与 Redis 快照字段。
- `device-admin-event-tree-ui`: 管理端事件树形列表、新增根/子事件及独立 logo/color 配置。
- `voice-event-child-disambiguation`: voice 侧子优先匹配、非叶子追问 pending 状态机与叶子落库规则。

### Modified Capabilities

- `device-event-cache-rebuild-on-mutate`: 缓存重建字段集增加 `parent_id`。
- `device-admin-event-logo-color-ui`: 列表由扁平改为树形，行内操作增加「新增子事件」。

## Impact

- **数据库**：`ai_voice_device.event` 需已有 `parent_id`；`child_ids` 不再使用。
- **device-service**：`internal/services/device/admin.go`、`cache_rebuild.go`、`eventListFields`；`api/v1/device_admin_http.go`；`internal/controller/device_admin_event.go`；契约 `DeviceAdminContract`。
- **voice-service**：`internal/services/voice/voice_chat_understanding.go`（匹配、pending、动作写库分支）；可选 DeepSeek 提示词补充。
- **前端**：`resource/public/admin.html` 事件管理模块。
- **entity/dao**：`ParentId` 保留；`ChildIds` 代码层忽略或后续从 entity 移除注释引用。

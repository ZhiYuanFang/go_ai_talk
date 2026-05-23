## Context

`event` 表当前为扁平字典：`ListEvents`、Redis 缓存与 `admin.html` 均以 id 升序平铺展示。voice 侧 `extractEventFromText` 无序遍历全部事件名，命中即写入 `history.event_id`。业务需要在「换尿布」等父事件下配置子事件（如大便/小便），且用户仅说父名时必须多轮追问，仅叶子节点可落库。

数据库已增加 `parent_id` 列；`child_ids` 不再作为设计的一部分。变更跨 **device-service**（权威数据源 + 缓存）、**admin.html**（管理 UI）、**voice-service**（匹配与 pending 状态机），须遵守服务边界：voice 经 `DeviceAdmin()` HTTP 读事件，不直连 device 库。

## Goals / Non-Goals

**Goals:**

- 以 `parent_id`（`0` = 根）表达**通用树**；子列表由 `WHERE parent_id = ?` 推导，不读写 `child_ids`。
- device API：新增/列表/删除/缓存规则与层级一致；同父下 `name` 唯一。
- 管理端树形展示；支持「新增子事件」；每个节点独立配置 `logo` / `color`（**不继承**父节点）。
- voice：子节点匹配优先；命中非叶子 → 内存 pending + 动态追问直接子节点名；仅叶子写 history；`finishTalk=false` 等待下一轮。
- pending 为进程内存态；会话 TTL 过期后不恢复（用户久后再说视为新轮次）。

**Non-Goals:**

- 历史 `history` 行迁移或回填。
- pending 写入 Redis / 与 session 同 TTL 持久化。
- 管理端拖拽改父级、防环可视化编辑器（第一期不做移动节点）。
- App 端事件选择器 UI（若存在）的树形改造（除非已有明确入口）。

## Decisions

### 1. 仅用 `parent_id` 邻接表，放弃 `child_ids`

**选择**：`parent_id` 为唯一层级字段；代码与 API 忽略 `child_ids`。

**理由**：`hasChildren(id)` 等价于 `EXISTS (parent_id = id)`；双写易漂移。

**备选**：保留 `child_ids` 冗余 — 拒绝。

### 2. 名称唯一性：同父下唯一

**选择**：`(parent_id, name)` 逻辑唯一；不同父下允许同名。

**理由**：通用树下「其他」等子名可能重复；全局唯一过严。

**实现**：`AddEvent` / `UpdateEvent` 查重时增加 `parent_id` 条件；DB 层可选唯一索引 `(parent_id, name)`。

### 3. 删除：有子则拒绝

**选择**：`DeleteEvent` 若存在 `parent_id = id` 的行则返回业务错误。

**理由**：避免孤儿节点；比级联删更安全。

### 4. ListEvents 仍返回扁平数组

**选择**：API 返回 `[]Event`（含 `parentId`），前端 `buildTree()` 渲染。

**理由**：与现有 consumers（voice、history 缓存）兼容；voice 自行建 `childrenByParent` 索引。

### 5. 缓存字段集扩展

**选择**：`eventListFields()` 与 `RebuildEventCache` 扫描增加 `parent_id`。

**理由**：voice/history 读缓存时需组树，缺字段则二次查库。

### 6. 管理端子事件不继承 logo/color

**选择**：`AddEvent(parentId>0)` 时 **不** 复制父 `logo`/`color`；表单使用与根事件相同的默认值。

**理由**：产品要求中间节点与叶子均有独立视觉配置。

### 7. Voice 匹配：深度 + 名称长度优先，pending 限定子树

**选择**：

1. 从 `ListEvents` 构建 `childrenByParent`、`depthById`。
2. 无 pending 时：全树匹配，排序为 **深度降序、名称长度降序**（叶子/深层优先）。
3. 命中事件且 `hasChildren` → `setPendingChild(deviceNo, {ParentEventId, ActionContext, ...})`，回复「{父名}是{子1}还是{子2}？」，`finishTalk=false`，**不写 history**。
4. 有 pending 时：**仅在** `childrenByParent[pending.ParentEventId]` 内匹配；命中仍为非叶子则 pending 下钻；命中叶子则按原动作（start/end/one）写 history 并 `clearPending`。
5. pending 存 `VoiceService` 内存 map（与 `pendingQuantity` 同模式），不持久化。

**理由**：复用现有数量追问的 `finishTalk` 语义；通用树支持多级下钻。

**备选**：DeepSeek 单独判断是否有子 — 拒绝为主路径，模型仅作兜底抽取。

### 8. DeepSeek 提示词（轻量增强）

**选择**：事件列表仍可扁平拼接，补充说明「存在 parent_id 层级；若用户仅提及父事件名且未指定子类型，返回该父事件名」；**是否追问由代码根据 hasChildren 决定**，不仅依赖模型。

### 9. `ChildIds` entity 字段

**选择**：entity 生成层若仍含 `ChildIds`，业务代码不读写；可选在后续 gf gen 后自然消失。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 父子名重叠导致误匹配 | 子优先 + 名长优先排序；pending 限定子树 |
| pending 多实例 voice 丢失 | 与现网 quantity pending 一致；可接受 |
| 同父同名并发插入 | DB 唯一索引 + API 查重 |
| 修改 parent_id 造成环 | 第一期禁止 API 修改 parent_id |
| 中间节点无 logo 管理端难辨认 | 占位 UI 与根事件一致，管理员自行上传 |

## Migration Plan

1. 确认 `event.parent_id` 已存在，默认 `0`；不处理 `child_ids` 列数据。
2. 部署 device-service（API + 缓存字段）→ 重建 Redis 事件缓存（或等待下次 mutate 触发）。
3. 部署 voice-service（匹配 + pending）。
4. 更新 `admin.html`（静态资源，随 gateway/device 发布）。
5. **回滚**：回退 voice/device 二进制；DB `parent_id` 可保留（旧代码忽略）；缓存自动随旧 Fields 重建（无 parent_id 字段时 JSON omitempty）。

## Open Questions

- 无（探索阶段已确认：必须追问、通用树、不继承 logo/color、pending 不持久化、不迁历史）。

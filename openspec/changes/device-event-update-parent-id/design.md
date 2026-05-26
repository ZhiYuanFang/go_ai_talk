## Context

- 基线：`openspec/changes/device-event-hierarchy` 约定 `parent_id` 邻接表、`AddEvent(parentId)`、删除有子拒绝、**UpdateEvent 不改 parent_id**。
- 现网：`UpdateEvent` 仅更新 name/type/extraNames/color/logo；`admin.html` 编辑模式不提交 `parentId`。
- 约束：device-service 独占 `event` 表；voice 经 HTTP 读 `ListEvents`，父节点变更后依赖现有 **mutate → RebuildEventCache** 路径。

## Goals / Non-Goals

**Goals**

- 管理员在**编辑已有事件**时可改父节点（含改为根 `parentId=0`）。
- 服务端保证树仍为 DAG（无环）、父存在、同父 `name` 唯一。
- 与现有删除/新增规则一致的错误码与中文提示。

**Non-Goals**

- 不提供「连同子树一起移动」的专用 API（子节点 `parent_id` 仍指向原父 id，仅被移动节点改父；若业务要整棵子树搬迁需另开需求）。
- 不修改 voice pending 持久化策略。

## Decisions

### 1. 在 `UpdateEvent` 增加 `parentID` 参数

**选择**：`UpdateEvent(ctx, id, ..., parentID int64)`；multipart 表单 `parentId` 可选，缺省表示**不修改**父节点（与「显式传 0 表示升为根」区分：编辑表单始终提交当前选择值，含 `0`）。

**理由**：避免误将未传字段当成 0 导致批量误提根。

**实现要点**：控制器在 update 路径解析 `parentId`；若表单未出现该字段且需兼容旧客户端，可约定仅在新版 admin 提交；本期 admin 与 API 同步上线，**编辑提交始终带 `parentId`**。

### 2. 成环检测

**选择**：当 `newParentID > 0` 时：

1. 拒绝 `newParentID == id`。
2. 在内存中由 `ListEvents` 或单次查询构建 `parentByID` / `childrenByParent`，收集 **`id` 的全部后代 id 集合**；若 `newParentID` 属于该集合则拒绝（不能把父设为自己的子孙）。
3. 校验 `newParentID` 对应行存在。

**理由**：事件字典规模小，全表扫描可接受；无需 DB 递归 CTE。

**备选**：仅向上爬 `newParentID` 祖先链是否含 `id`——与向下收集后代等价，实现取一种即可。

### 3. 同父唯一与父变更

**选择**：更新时查重使用 **`newParentID`**（若本次修改父）或 **现有 `parent_id`**（若父不变）与 `name` 做 `eventNameExistsUnderParent`，排除自身 `id`。

### 4. 有子节点的事件是否允许改父

**选择**：**允许**。移动父节点不改变子行的 `parent_id`（子仍挂在该节点下），树局部形状合法。仅禁止成环与无效父 id。

**理由**：与「换尿布」整棵子树挂在某中间节点下的模型一致；运营常只需把叶子或中间节点换父。

### 5. 管理端父事件选择器

**选择**：编辑弹窗增加下拉（或搜索列表），选项为**除自身及其后代外的**全部事件 +「无（根）」；默认选中当前 `parentId`。

**理由**：防止前端提交必然后端拒绝的成环请求，减少误操作。

### 6. 错误语义

**选择**：新增或复用 device 包 sentinel：

- 父 id 不存在 → 参数/未找到类错误（与 AddEvent 一致）。
- 成环 / 自身为父 → `CodeInvalidOperation` + 中文「不能将父事件设为自己或其子事件」。
- 同父同名 → 现有 `ErrEventExists`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 运营误把中间节点提根导致 voice 追问范围变化 | 管理端文案提示；voice 按最新 `ListEvents` 建树 |
| 并发改父与同父同名 | 保持 DB 唯一索引 `(parent_id, name)` + API 查重 |
| 旧 admin 缓存未带 parentId 提交 | 与本变更同发静态资源 |

## Migration Plan

1. 部署 **device-service**（UpdateEvent + 错误码）。
2. 发布 **admin.html**（编辑带 parentId）。
3. 任意一次 event mutate 触发 Redis 全量重建；或管理端手动改一条事件触发。
4. **回滚**：回退二进制；已改的 `parent_id` 保留，旧代码忽略新父关系直至再次升级。

## Open Questions

- 无（默认：编辑表单**必须**显式提交 `parentId`，包括 `0`）。

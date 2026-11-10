## Why

Flutter 预测算法需要按事件根决定「是否把子树历史合并成一条预测」。该开关属于全局事件字典配置，目前库表与 Hub 均无字段可运维勾选，客户端只能硬编码或全量不合并。需在事件主档增加 `merge_predict`，默认否，仅「有子节点的根事件」可开启，供 Flutter 经既有 options/list 读取。

## What Changes

- 全局事件字典 `event` 增加列 `merge_predict`（`TINYINT NOT NULL DEFAULT 0`）；对外 JSON 字段 `mergePredict`（与 `isAppointment` 同 camelCase 惯例）。
- Hub 事件管理：列表展示、表单勾选；**仅当**该事件为根（`parent_id=0`）**且**存在至少一个直接子节点时允许勾选；叶子与无子根禁用勾选。
- 服务端写路径（`AddEvent` / `UpdateEvent`）强制校验：不满足「有子的根」时 MUST 将标志置为 `0`（防表单/API 绕过）；根改挂为非根、或删光子节点后更新时 MUST 清零。
- 既有事件 options / list / 内部字典返回路径附带 `mergePredict`；写库成功后走既有 EventOptions 缓存重建。
- **不改** voice/history 预测推送、临近 pending、准确率聚合等服务端算法；本仓不实现 Flutter 预测逻辑。

## Capabilities

### New Capabilities

- `event-merge-predict`：事件字典「是否合并预测」标志的存储、根+有子约束、Hub 编辑与对 Flutter 可读的 options/list 暴露。

### Modified Capabilities

- （无）——不修改 `predict-imminent` / Live Activity / Hub 准确率等同根聚合的既有需求；复用 v3.0.3 中事件层级（`parent_id`）与 EventOptions 写后重建约定，不改其 Requirement 文本。

## Impact

- **device-service / device DB**：`event` 表新列（启动 ensure 幂等）；DAO/entity/do；admin 增改与列表字段；`Ensure*` 启动钩子。
- **Hub**：`resource/public/admin.html` 事件表列与表单 checkbox（条件启用）。
- **读模型**：凡返回 `entity.Event` 的 App/内部 options 与后台 list 自然带上新字段（旧 Redis 快照随下次写后重建淘汰）。
- **history / voice**：无算法变更；经契约拿到的事件项可含字段但本变更不要求其消费。
- **Flutter 孪生仓**：消费 `mergePredict` 做预测算法（本 change 任务以本仓后端+Hub 为准）。
- **API**：options/list **可选加字段**，非 BREAKING；无新 App 路由，无 usage 统计变更。
- **重部署**：`device-service`（schema ensure + admin）；`gateway-app-server`（静态 Hub 页）。不必重部署 voice/history（除非仅因共用镜像策略统一发版）。

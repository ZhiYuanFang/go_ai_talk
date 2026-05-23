## Context

- 行内色调：点击表格「色调」列 → 设置 `pendingInlineEventRow` → 触发 `#eventInlineColor` 的 `click()` → 监听 `change` 立即 `submitEventRowUpdate`。
- 更新仍走 `POST /device/admin/api/event/update`（multipart），需携带整行 `name`、`needQuantity`、`extraNames`、`color`；未传 logo 文件则保留原 logo。

## Goals / Non-Goals

**Goals:**

- 用户选色后可预览，点击 **确定** 才提交；**取消** 放弃本次修改。
- 交互与现有 `eventMsg` 错误提示、行级 `event-row-busy` 状态兼容。

**Non-Goals:**

- 不改 Logo 行内「选文件即提交」。
- 不改 `device-service` 事件 Redis 缓存刷新（`ListEvents` 读旧缓存问题）。
- 不替换为第三方重型颜色选择组件库。

## Decisions

### 1. UI：行内浮层（Popover）

点击色调单元格时：

1. 记录 `pendingInlineEventRow` 与锚点元素（`td`）。
2. 显示 `#eventColorPopover`（绝对/fixed 定位在单元格附近）：
   - `<input type="color" id="eventInlineColorPick">`（可见，非仅 hidden trigger）
   - 只读或可编辑文本展示 `#RRGGBB`（可选 `input type="text"` 同步）
   - 按钮：**确定**、**取消**
3. **确定**：取当前色值 → `submitEventRowUpdate(row, { color })` → 成功关闭浮层。
4. **取消** 或点击浮层外（可选）：关闭浮层，不提交。

**弃用**：隐藏 `input` + 单独 `change` 监听即时提交。

### 2. 色值格式

- 提交前 `normalizeEventColor`（`#RRGGBB`），与后端 `ValidateEventColor` 一致。
- 浮层打开时用当前行 `row.color` 初始化取色器。

### 3. 并发与关闭

- 打开新行色调前关闭已有浮层。
- 提交中禁用浮层按钮，与 `eventRowBusyId` 一致。

### 4. 无障碍与移动端

- 确定/取消使用 `<button type="button">`，避免误提交表单。
- 小屏下浮层可 `position: fixed` 居中，避免被表格 `overflow` 裁切（实现时检测或统一 fixed + 靠近点击坐标）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 表格 `overflow-x` 裁切浮层 | 使用 `fixed` 定位或挂到 `body` |
| 多一次点击 | 符合「需确定按钮」的产品要求 |

## Migration Plan

- 仅替换静态 `admin.html`；部署 gateway 后强刷管理页即可，无需改服务二进制（若静态由 gateway ServeFile 提供则重建 gateway）。

## Open Questions

- 无。

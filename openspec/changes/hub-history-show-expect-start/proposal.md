## Why

运维在 Hub 看历史事件时只能看到实际开始/结束，无法对照「这一次原本预计何时开始」。list 接口已回传可选 `expectStartTime`，缺的是表上展示。使用统计右侧嵌入历史 Tab 刚对齐主表列，需同步加列以免两处不一致。

## What Changes

- **`history.html` 历史事件表**：增加只读列 **「预计开始」**，展示行内 `expectStartTime`（Unix 秒格式化；空/缺省显示 `-` 或等价占位）。
- **使用统计右侧历史 Tab**：`api-usage-stats.html` / `client-usage-stats.html` 同步同一列与同一格式化语义。
- **非目标**：不改 history-service / admin list API；不展示预计结束/预计数量（本变更仅预计开始）；不改预测准确率卡；不改 Flutter。

## Capabilities

### New Capabilities

- `hub-history-show-expect-start`：Hub 设备历史表与使用统计嵌入历史表展示 `expectStartTime`。

### Modified Capabilities

- （无）

## Impact

- **静态页**：`resource/public/history.html`、`api-usage-stats.html`、`client-usage-stats.html`
- **gateway-app**：静态资源发版
- **history-service**：无（依赖既有 list 已回传字段）

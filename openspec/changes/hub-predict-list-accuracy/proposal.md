## Why

运维在 Hub「设备数据运维」页只能看 Redis 实时预测上报与原始 history，看不到 Flutter 预测竖列表那套「近 7 日准确率柱 + 点日小结」。`history-expect-fields-*` 已落库 `expectStartTime`，Flutter `predict-list-accuracy-bars` 已定计分规则；Hub 需同规则只读展示，便于核对预报准不准。

## What Changes

- 在 `history.html`（按 `deviceNo`）内嵌**仿预测竖列表**区域：按事件根（或可解析的根）纵向卡片；每卡近 7 自然日准确率柱（柱高 0–100%，柱顶整数 `%`）；无合格样本日不画柱、不标 `0%`。
- 点击有样本的日柱：右侧（宽屏）或卡下展示该日**预测小结**（预计开始 / 实际开始 / 是否命中）；提供返回「近7日」。
- 新增 history **Admin 只读 API**：按 `deviceNo` 返回近 7 日按事件聚合的日准确率与可选日明细；计分对齐 Flutter：`|actualStart−expectStartTime|≤1800s`，仅已存正 `expectStartTime`，禁止重放。
- **不改** App PUT、Redis 预测上报卡、Flutter UI；不计 duration/amount 维。
- **非目标**：不做新 ticker/Redis 准确率缓存；不改趋势中心；不发推送。

## Capabilities

### New Capabilities

- `hub-predict-list-accuracy`：Hub 设备页纵向准确率列表（Admin API + `history.html` 展示），计分与 Flutter `predict-list-accuracy-bars` / `trend_predict_accuracy.dart` 对齐。

### Modified Capabilities

- （无）

## Impact

- **history-service**：Admin GET（鉴权同既有 history Admin）；读 history 表近窗行；可选经 `clients/device` 解析事件根以便与 App 同卡聚合。
- **gateway-app**：`/device/admin/api/history/*` 反代通常已覆盖；须部署更新后的 `history.html`。
- **Hub 静态**：`resource/public/history.html` + `admin-pages.css` 左右分栏约定。
- **依赖**：`history-expect-fields-go`（列已存在）；Flutter twin 规则权威 `kPredictAccuracyHitWindowSec=1800`。
- **usage**：Admin 非 App 接口，不改 `maintenance_skip`。

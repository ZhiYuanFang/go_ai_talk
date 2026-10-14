## Why

运维在 Hub「客户端使用统计 / App API 使用统计」上常用较短窗口查看近期活跃，但时间范围下拉仅有近 7 / 30 天，默认 7 天偏宽。后端 `days` 查询本已支持 1–30，只需补齐前端选项并默认近 1 天，无需改排序或拉全量。

## What Changes

- `client-usage-stats.html` 与 `api-usage-stats.html` 时间范围下拉增加「近 1 天」「近 3 天」。
- 两页默认选中「近 1 天」（原默认近 7 天）。
- **不**改 wx 列表排序、不新增自填天数、不改后端 `NormalizeDays` / Redis 窗口上限、不改 API 契约。

## Capabilities

### New Capabilities

- `hub-usage-stats-days-presets`：两页统计时间范围预设含 1/3/7/30 天，默认 1 天。

### Modified Capabilities

- （无。基线无已归档的独立 Hub 天数预设能力；后端 days 语义不变。）

## Impact

- **静态页**：`resource/public/client-usage-stats.html`、`resource/public/api-usage-stats.html`。
- **gateway-app**：发版带静态资源即可。
- **voice / device / Python**：无。
- **usage 统计**：Admin 页，不改 `maintenance_skip`。

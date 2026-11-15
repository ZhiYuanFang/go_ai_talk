## Why

预测 Pro 想法墙 submit 已落库，但 App list 与 Hub admin list 均返回空列表。根因是查询使用 `WhereNull("deleted_at")`（仅 `IS NULL` 视为未删），而线上新增行的 `deleted_at` 实际为默认值 `0`，导致全部活跃想法被滤掉。

## What Changes

- 统一 `predict_pro_idea` 软删「未删」语义：`deleted_at = 0` 或 `NULL` 均为未删；`deleted_at > 0` 为已软删。
- App list / Admin list / loadActive / 更新展示文案 / 软删前置校验等全部查询条件对齐上述语义（不再仅依赖 `IS NULL`）。
- EnsureSchema / DDL 注释与默认值与线上一致（未删默认 `0`），避免新环境再出现 NULL/0 分裂。
- 不改 App/Admin HTTP 路径与响应字段结构；不改日限额（仍含软删行 COUNT）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-pro-idea-data-model`: 明确 `deleted_at` 未删值为 `0`（兼容存量 `NULL`），软删写入 unix 秒。
- `predict-pro-idea-api`: App/Admin 列表与按 id 加载活跃想法 MUST 将 `deleted_at=0`（及 `NULL`）视为可见/可操作。

## Impact

- 代码：`internal/services/device/predict_pro_idea.go`、`schema_predict_pro_idea.go`
- 进程：仅 `device-service`（Ensure + 查询语义）；gateway / App 契约路径不变
- 数据：存量 `deleted_at=0` 行在修复后立即出现在 list/Hub，无需洗数；已软删（`>0`）仍不可见
- 不必改 Flutter / Hub 前端（只要后端 list 有数据即可）

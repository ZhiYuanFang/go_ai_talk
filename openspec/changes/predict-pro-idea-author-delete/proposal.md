## Why

Hub Admin 已有 `POST /device/admin/api/predict-pro-ideas/delete` 软删，但 App 端作者无法自删。Flutter 产品要做「长按本人想法 → Sheet → 删除」，需对称的 App 接口；与姊妹仓 `flutter_ai_talk` 同名提案同步，**本仓（Go）先实现并部署，Flutter 后对接**。

## What Changes

- 新增 App：`POST /device/app/api/predict-pro-ideas/delete`，body `{ "id": <ideaId> }`，登录会话。
- 仅当想法未删且 `wx_id` 等于调用者时软删（复用现有 `deleted_at` 语义：`>0` 为已删）；非作者 / 已删 / 不存在 → 业务错误。
- 软删后 list 不再返回该行；不改 Admin 删除路径；gateway 须可转发该 App 路径（若走既有 fuzzy 代理则核对即可）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-pro-idea-api`：增加作者自删 App 端点与授权语义。

## Impact

- 代码：`api/v1/device_app_predict_pro_idea_http.go`、`device_app_predict_pro_idea.go` controller、`predict_pro_idea.go` service（可扩展 `SoftDeletePredictProIdea` 或新增带 `wxId` 校验的方法）。
- 进程：`device-service`（必）；`gateway-app-server`（若代理/白名单需改则一并）。
- **姊妹提案（后做）**：`d:\work\flutter_ai_talk\openspec\changes\predict-pro-idea-author-delete\`。
- 顺序：本变更 apply/部署完成后，再通知 Flutter `/opsx-apply`。

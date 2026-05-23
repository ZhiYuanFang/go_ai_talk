# 设计：管理端问答库分页与全量页

## API

| 路径 | 方法 | 说明 |
|------|------|------|
| `/device/admin/api/qa/list` | GET | `page`（默认 1）、`pageSize`（默认 10，最大 100）→ `{ list, total, page, pageSize }` |
| `/device/admin/api/qa/delete` | POST | `{ id }`，需 `X-Admin-Password` |
| `/voice/internal/api/qa/list` | GET | 同上，voice 库权威 |
| `/voice/internal/api/qa/delete` | POST | `{ id }` |

## 排序与默认值

- 查询：`OrderDesc(id)`，较新记录在前。
- `pageSize <= 0` 时视为 10；超过 100 截断为 100。

## 服务边界

- `internal/services/voice/qa_internal.go` 使用 `dao.Qa`（本域表）。
- `internal/services/device` 经 `voice_upstream_qa.go` HTTP 调用，禁止 import `dao.Qa`。

## 前端

- `admin.html`：`loadQaList` 请求 `page=1&pageSize=10`；`total > 10` 显示链接 `/device/admin/qa-records`。
- `qa-records.html`：口令登录、每页 20 条、上一页/下一页、行内删除。

## 静态路由

- `register.go`（device-service）与 `gateway_app_register.go` 绑定 `/device/admin/qa-records` → `resource/public/qa-records.html`。

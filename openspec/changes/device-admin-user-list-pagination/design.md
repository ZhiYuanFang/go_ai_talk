# 设计：设备记录分页与最近 HTTP 接口

## 数据模型

`user` 表新增：

| 列 | 类型 | 说明 |
|----|------|------|
| `last_api_path` | VARCHAR(256) | 最近一次对外 HTTP，格式 `METHOD /path`（无 query） |
| `last_api_at` | BIGINT | Unix 秒 |

与 `last_talk_*` 并列，不互相替代。

DDL 参考：`hack/ddl_device_user_last_api.sql`。

## API

| 路径 | 方法 | 说明 |
|------|------|------|
| `/device/admin/api/user/list` | GET | `page`（默认 1）、`pageSize`（默认 5，最大 100）、`q`（可选模糊）→ `{ list, total, page, pageSize }` |
| `/device/internal/api/user/touch-api-access` | POST | `{ deviceNo, apiPath, at }`，仅服务间/网关调用 |

排序：`last_api_at DESC, id DESC`（未 touch 的设备 `last_api_at=0` 靠后）。

## 网关 touch middleware

- 挂载于主网关与 `gateway-app-server` 的 `/*`（在 Bearer 注入之后执行于 gateway-app）。
- 跳过：WebSocket Upgrade、`/device/internal/` 前缀、无法解析 `deviceNo` 的请求。
- `deviceNo` 解析顺序：query `deviceNo` → `X-Internal-Device-No` → `X-Device-No` → JSON body `deviceNo`（POST，`Content-Type` 含 json；读 body 后还原供反代）。
- 记录路径：`METHOD + " " + URL.Path`。
- 异步调用 device internal touch（失败静默，不阻塞主请求）。

## 前端

- `mainCard` 仅保留注册设备。
- 新 `#deviceRecordCard`：搜索框、分页（每页 5）、表格含 `last_api_path` / `last_api_at`；`device_no` 为 `/device/history/{deviceNo}` 链接。

## 服务边界

- touch 与列表读写均在 `internal/services/device` + `dao.User`（device 库）。
- 网关不直连 device 库表。

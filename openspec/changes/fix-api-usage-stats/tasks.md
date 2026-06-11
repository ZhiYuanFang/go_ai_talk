## 1. 修复 Redis 读取

- [x] 1.1 `usagestats/store.go`：`redisHashToMap` / `redisString` 支持 `*gvar.Var`，HGET 单值正确解析

## 2. 修复统计写入（反代路径）

- [x] 2.1 `domain_route_proxy.go`：`buildReverseProxy.ModifyResponse` 对 2xx 调用 `RecordHTTPRequest`
- [x] 2.2 `record_http.go` + middleware：`RecordGHTTPRequest` 覆盖本机 Handler

## 3. wx 账号列表 API

- [x] 3.1 `api/v1` 定义 `GET /device/admin/api/wx/list`
- [x] 3.2 device-service 实现分页查询 wx 表
- [x] 3.3 gateway device 反代已覆盖 `/device/admin/api/*`

## 4. 前端 api-usage-stats.html

- [x] 4.1 「按用户」Tab：wx 列表 + 点选加载 usage/user
- [x] 4.2 空态与 hint 文案（不含 WS、登录前 wxId=0 等）

## 5. 维护型 API denylist

- [x] 5.1 `maintenance_skip.go`：token/refresh、version/check、site/home、version/admin/*
- [x] 5.2 接入 `ShouldSkipRecord` 与 `ShouldSkipHTTPRecord`

## 6. 读 API 排序

- [x] 6.1 list/detail/user 支持 `sortBy=count|lastAt`，默认 count ↓
- [x] 6.2 前端排序下拉与 query 传参

## 7. 校验

- [x] 7.1 `openspec validate fix-api-usage-stats --strict`
- [x] 7.2 `go build ./...`

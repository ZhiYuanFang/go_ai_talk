## 1. API 层 - 新增请求响应结构体

- [x] 1.1 在 `api/v1/device_history_http.go` 新增 `DeviceHistoryFilterReq` 结构体（路径 `GET /device/history/api/filter`，字段：deviceNo、eventIds、startTime、endTime、limit）
- [x] 1.2 在 `api/v1/device_history_http.go` 新增 `DeviceHistoryFilterRes` 结构体（字段：`List []entity.History`）
- [x] 1.3 新建 `api/v2/device_history_http.go`，新增 v2 `DeviceHistoryListReq` 结构体（路径 `GET /device/history/api/v2/list`，字段：deviceNo、page、pageSize、startTime、endTime、limit）
- [x] 1.4 在 `api/v2/device_history_http.go` 新增 v2 `DeviceHistoryListRes` 结构体（与 v1 同结构：list、total、page、pageSize）

## 2. 契约层 - 扩展服务接口

- [x] 2.1 在 `internal/services/contracts/runtime_contracts.go` 的 `DeviceHistoryContract` 接口新增 `ListHistoryFilter` 方法签名
- [x] 2.2 在 `internal/services/contracts/runtime_contracts.go` 的 `DeviceHistoryContract` 接口新增 `ListHistoryPageV2` 方法签名
- [x] 2.3 在 `internal/services/contracts/http_targets.go` 新增 `HistoryFilterPath()` 方法（返回 `/device/history/api/filter`）
- [x] 2.4 在 `internal/services/contracts/http_targets.go` 新增 `HistoryListV2Path()` 方法（返回 `/device/history/api/v2/list`）

## 3. 服务层 - local 实现

- [x] 3.1 在 `internal/services/history/local.go` 的 `localService` 结构体新增 `ListHistoryFilter` 方法（委托给包级函数）
- [x] 3.2 在 `internal/services/history/local.go` 的 `localService` 结构体新增 `ListHistoryPageV2` 方法（委托给包级函数）
- [x] 3.3 在 `internal/services/history/local.go` 实现 `ListDeviceHistoryFilter` 包级函数：动态拼接 WHERE 条件（deviceNo、eventIds IN、startTime >=、endTime <=），ORDER BY id DESC，Limit 限制，默认 100，上限 500
- [x] 3.4 在 `internal/services/history/local.go` 实现 `ListDeviceHistoryPageV2` 包级函数：支持 limit > 0 时替代 pageSize（page 固定为 1），支持 startTime/endTime 时间过滤，COUNT 也带相同 WHERE 条件，pageSize 上限 100

## 4. 服务层 - adapter 实现（三处同步）

- [x] 4.1 在 `internal/services/history/adapter.go` 的 `historyRemoteClient` 结构体新增 `ListHistoryFilter` 方法（HTTP GET 调用远程 filter 接口）
- [x] 4.2 在 `internal/services/history/adapter.go` 的 `historyRemoteClient` 结构体新增 `ListHistoryPageV2` 方法（HTTP GET 调用远程 v2 list 接口）
- [x] 4.3 在 `internal/services/history/adapter.go` 的 `switchAdapter` 结构体新增 `ListHistoryFilter` 方法（支持 local/remote/canary 切换和 failover）
- [x] 4.4 在 `internal/services/history/adapter.go` 的 `switchAdapter` 结构体新增 `ListHistoryPageV2` 方法（支持 local/remote/canary 切换和 failover）

## 5. Controller 层

- [x] 5.1 在 `internal/controller/device_history.go` 新增 `Filter` 方法（参数校验：deviceNo 必填；解析 eventIds 逗号分隔字符串为 []int64；调用 `Svc.ListHistoryFilter`；返回 `DeviceHistoryFilterRes`）
- [x] 5.2 在 `internal/controller/device_history.go` 新增 `ListV2` 方法（参数校验：deviceNo 必填；page/pageSize 默认值逻辑同 v1；传递 startTime/endTime/limit；调用 `Svc.ListHistoryPageV2`；返回 `DeviceHistoryListRes`）

## 6. 验证

- [x] 6.1 运行 `go build ./...` 验证编译通过
- [x] 6.2 运行 `go vet ./...` 验证无静态检查错误

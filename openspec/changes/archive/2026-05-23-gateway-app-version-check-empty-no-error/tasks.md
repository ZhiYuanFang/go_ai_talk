## 1. 版本检查控制器

- [x] 1.1 `GatewayAppCtrl.VersionCheck`：DB 查询改为 **`One` + `IsEmpty`**（或等价），无行时直接返回 **`buildVersionRes`** 与「无配置」一致的成功体（**`needUpdate=false`**），**不得**因 `sql.ErrNoRows`/`Scan` 空集返回 `err != nil` 给 HTTP 层
- [x] 1.2 区分「无行」与真实 DB 错误：仅后者打 **Warning/返回错误**；空表路径避免 error 级日志误导

## 2. 校验

- [x] 2.1 `openspec validate gateway-app-version-check-empty-no-error`
- [x] 2.2 `go build ./...`

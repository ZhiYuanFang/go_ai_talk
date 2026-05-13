## Why

App 端「检查更新」依赖 `GET /device/app/api/version/check` 读取网关库中的版本配置表。当表内尚无任何版本记录时，部分驱动/ORM 路径会返回「无行」类错误，客户端收到非 0 业务码或异常信息，与「当前无可发布版本＝客户端无需更新」的产品语义不符。

## What Changes

- **`GET /device/app/api/version/check`**：当版本表**无数据**（或等价于「无可用最新版本行」）时，**不得**作为失败响应；应返回 **`code=0`**，且 **`needUpdate=false`**（无需更新），字段取值与「无配置」语义一致（见 design）。
- 区分**表空/无行**与**真实数据库故障**（连接失败、语法错误等）：后者仍可返回错误；前者归入成功 + 无需更新。
- 联调页/客户端：可继续依赖「成功 + needUpdate」判断，无需把空表当异常分支。

## Capabilities

### New Capabilities

- `gateway-app-version-check`：定义 App 网关版本检查在「无版本数据」时的成功响应语义，以及与 Redis 缓存、DAO 查询的配合边界。

### Modified Capabilities

- （无）仓库 `openspec/specs/` 下尚无已归档的同名能力；本变更为新增能力规格。

## Impact

- **代码**：`internal/controller/gateway_app_ctrl.go` 中 `VersionCheck` / `buildVersionRes` 及（若需要）DAO 查询方式（如 `One`+`IsEmpty` 与 `Scan` 空集错误处理）。
- **API**：响应 JSON 形态不变，仅保证空表路径下 **`code` 恒为成功** 且 **`needUpdate=false`**。
- **观测**：空表路径不应打「错误级」日志误导运维；真实 DB 错误保留告警。

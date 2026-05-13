## Context

- 实现入口：`GatewayAppCtrl.VersionCheck`（`GET /device/app/api/version/check`），读网关默认库 `app_version`（DAO：`dao.AppVersion`），并可读 Redis 缓存 `gw:app:version:latest`。
- 现状：已对 `Scan` 返回 `err != nil` 时返回「成功 + needUpdate=false」并打 Warning；但在部分环境下「表空」可能表现为 **`sql: no rows in result set`** 或其它 ORM 包装错误，仍被当作「读表失败」分支处理，或缓存反序列化失败后回落 DB 仍踩空行错误，客户端感知为失败。

## Goals / Non-Goals

**Goals:**

- 将「**无任何版本行**」明确归类为**业务成功**：`code=0`，**`needUpdate=false`**（无需更新），不向客户端返回错误码。
- 响应字段与现有 `GatewayAppVersionCheckRes` 对齐：`latestVersion` 在无配置时取**当前请求 `currentVersion`**（与现有 Scan 失败兜底一致），`releaseDate` 等数值字段为 0，字符串字段为空串，`forceUpdate` 为 false。
- 查询实现上优先使用 **`One` + `IsEmpty`**（或等价）区分「无行」与「真实 SQL 错误」，避免把 `ErrNoRows` 当故障。

**Non-Goals:**

- 不改变「有版本数据」时的比较与 `needUpdate` 判定逻辑。
- 不在本变更中引入新版本表 DDL 或管理端写库流程。

## Decisions

### D1：空表 = 成功 + 无需更新

- **决策**：无版本行时返回与「无可用最新版本」一致的**成功**载荷，**禁止**因空结果集返回 `code != 0`。
- **理由**：与产品「无可更新版本」一致，客户端无需特殊错误分支。

### D2：Redis 缓存损坏或过期后的 DB 回落

- **决策**：若 Redis 命中但反序列化失败或内容无效，回落 DB；DB 无行时仍走 D1，不打 error 级日志（可 debug）。
- **理由**：避免缓存脏数据导致用户侧失败。

### D3：真实数据库故障

- **决策**：连接失败、权限、语法错误等仍返回错误（`code != 0`），与「无行」区分。
- **理由**：运维需感知基础设施问题。

## Risks / Trade-offs

- **[Risk] 将「无行」与「读库失败」混判** → **缓解**：显式 `errors.Is(err, sql.ErrNoRows)`（若底层暴露）+ `One/IsEmpty` 主路径。
- **[Trade-off] latestVersion 回填 current** → 客户端 UI 可能显示「最新=当前」；与现网 Scan 错误兜底一致，评审接受即可。

## Migration Plan

1. 部署 gateway-app-server；无需数据迁移。
2. 可选：清空错误监控里由空表触发的旧告警阈值。

## Open Questions

- 无。

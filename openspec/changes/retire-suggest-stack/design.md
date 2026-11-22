# Design

## Context

- 生成路径已在既有 `clinic-bridge` 退役；本仓对 `dao.Suggest` 仅剩 List/Delete 与 history 改名误迁（见 proposal.md — Why）。
- 权威数据曾约定在 voice 库 `suggest`；history 经 HTTP 反代 voice internal。Hub `history.html` 仍有隐藏 `loadSuggest`。
- 未归档变更 `hub-rename-baby-device-no` 的 delta 仍写「迁 history **与** suggest」——与本变更冲突，归档前须以本变更语义为准（见 D5）。

## Goals / Non-Goals

**Goals:**

- 拆除 suggest 读删 HTTP 面与 Hub UI；编译与路由无残留正式入口。
- `MigrateDeviceNo` 仅迁有效分区表，缺 `suggest` 表不再 1146。
- 规格与实现一致：无写入、无运维列表。

**Non-Goals:**

- DROP/`TRUNCATE` 遗留表或数据回填。
- Hub 自由填写 OLD→NEW 补迁 UI（改名运维另案）。
- 兄弟仓 Flutter/Python 深度清理（本仓保证 Go 无入口即可）。
- 改写统一意图主路径架构；仅保证 `target_type=suggest` 不落库、不走已删生成函数。

## Decisions

### D1 — 整链删除读删面，而非 410 占位

- 删除 `api/v1` suggest 请求类型、voice/history/admin controllers、`ListSuggest`/`DeleteSuggest` 契约方法、HTTP path helpers、delegate/adapter/client 实现、`suggest_internal.go`、register bind。
- Hub：去掉成长建议区块、文案与 `loadSuggest` 及调用点。
- **备选**：路由保留返回空列表 — 拒绝（继续暗示产品存在）。

### D2 — 改名迁移去 suggest

- `internal/services/history/migrate_device_no.go` 去掉对 `dao.Suggest` 的 UPDATE；内部 migrate API/注释同步。
- 不在 voice 库另做 suggest 改挂（遗留行可残留旧号，无产品读取）。

### D3 — DAO / entity 处理

- 应用代码零引用后：删除或停止生成对 `Suggest` 的业务依赖（`entity`/`do`/`dao` 与 `hack` dao 配置中的表项，以实现期编译干净为准）。
- 物理表不在本变更 DROP。

### D4 — `target_type=suggest` 字面

- 解析枚举可保留字符串兼容（避免未知类型 panic），业务分支 MUST NOT 写 suggest / 调已删生成。
- 与 conversation 同等「仅回复」或忽略专用处理即可；不新增产品能力。

### D5 — 与 `hub-rename-baby-device-no` 制品

- 实现本变更时同步修正该未归档 change 的 design/spec/tasks 中「迁 suggest」表述，或归档时以本 delta 覆盖主规格中的迁移条款，避免双源真相。
- 本变更 **不** 负责 Hub 半失败自由号补迁。

## Risks / Trade-offs

- [外部仍调旧 path] → 无路由/404；可接受，产品已无数据。
- [遗留 suggest 行挂旧号] → 无读面；若日后 DROP 表另开变更。
- [hub-rename 未归档 delta 仍含 suggest] → D5 修正制品或归档顺序以本变更为准。
- [history 进程曾用 `dao.Suggest` 跨域迁表] → 删除该 UPDATE 同时修复包边界违规。

## Migration Plan

1. 合并本变更代码；重部署 **voice-service**、**history-service**（及提供 Hub/admin history 反代的 gateway/device 面若静态资源随其发布）。
2. Hub 静态 `history.html` 随托管该资源的进程一起发布。
3. 回滚：恢复路由与 migrate suggest 块（不推荐；无产品价值）。
4. 半失败改名环境：去 suggest 后可重试 history/cash 段（仍可能需另案的自由号入口）。

## Open Questions

（无阻塞项；DROP 表时机留给后续运维变更。）

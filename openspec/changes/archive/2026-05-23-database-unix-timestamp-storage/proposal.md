## Why

当前部分库表与领域模型将时间以「可读的日期时间字符串」或与本地时区强绑定的形式落库/传输，海外多时区场景下易出现解释歧义、排序/区间查询不一致及夏令时边缘问题。统一改为**数字时间戳（明确纪元与单位）**存储与契约，可在不改业务语义的前提下为后续国际化预留一致的时间语义。

## What Changes

- 建立全库（或按服务边界内的库）**时间类字段**的存储规范：**Unix 秒 `BIGINT`**（权威单位），禁止将「本地日历字符串」作为权威落库形态。
- **BREAKING**：涉及表结构、DAO/entity、对外 HTTP/JSON 字段类型或格式的迁移；旧数据需可脚本化迁移并具备回滚或双写窗口策略（由 design/tasks 细化）。
- API 与内部 DTO：对外暴露数字时间戳时须在 OpenAPI/接口文档中**标明单位为秒**；迁移期若仍为字符串，须在变更说明中标为过渡并与 runbook 一致。

## Capabilities

### New Capabilities

- `database-unix-timestamp-storage`：约定纪元（Unix）、单位（**秒**）、MySQL 列类型、与 `time.Time`/JSON 的映射规则；迁移与兼容策略；服务边界内各表清单与验收要求。

### Modified Capabilities

- （无）当前 `openspec/specs/` 下无专门约束「库表时间列形态」的既有能力；本变更以新增能力规格为主。若实施阶段发现与 `history-service-db-ownership` 等文档冲突，再在对应 spec 增加 delta。

## Impact

- **数据库**：`manifest`/`hack`/`sql` 或各服务迁移脚本中的 DDL；可能涉及 `ai_voice_*` 多库。
- **代码**：`internal/model/entity`、`internal/model/do`、`internal/dao`（多为 gen 产物需重 gen 或手工对齐）、各 service 读写时间与序列化逻辑。
- **契约**：`api/v1` 中时间字段、前端联调页与网关集成测试（若有）。
- **运维**：发布顺序、数据回填窗口、监控与校验（空值、非法戳）。

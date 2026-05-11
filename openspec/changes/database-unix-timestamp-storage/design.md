## Context

- 业务已存在「时间戳」命名的字段仍以 `string` 等在应用层承载（如 `history.start_time`），部分字段注释仍写 UTF8 日期串，与「纯数字、时区无关」目标不一致。
- 多进程（device / history / voice / worker / gateway-app）分库边界已由仓库规则约束；时间列改造须在**各自服务所属库**内完成迁移与回归，禁止跨库直查。
- `internal/dao`、`internal/model/entity` 等多为 GoFrame gen 产物，DDL 与代码需**同源变更**（gen 重跑或受控手工对齐）。

## Goals / Non-Goals

**Goals:**

- 规定全项目权威落库的**时间语义**：Unix 纪元 + **秒**（`BIGINT`），全链路同一单位，避免与毫秒混读。
- 为每类「当前为字符串日期/模糊时间」的列给出：目标类型、回填规则、空值/非法值处理、对外 JSON 字段名是否保持不变。
- 给出可执行的**迁移顺序**（先加列/双写 → 回填 → 切换读 → 删旧列）与回滚要点。

**Non-Goals:**

- 不在本变更中规定各国家本地化**展示**格式（展示仍可在网关/App 层用 IANA 时区渲染）。
- 不引入新的全局「业务日历」类型（如仅 `DATE` 无时间的生日业务规则若有特殊语义，在业务层单独建模，本设计只约束「时刻」类字段）。

## Decisions

1. **存储单位：Unix 秒 `BIGINT`**
   - **理由**：与现网 `version.release_date` 等已有秒级 bigint 一致；列更窄；业务无亚秒权威落库需求。
   - **备选**：Unix 毫秒 — 已否决，避免与秒级存量列混用。
2. **语义：UTC 纪元，不包含「本地墙钟」**
   - **理由**：落库值与时区无关，区域展示由客户端或服务在已知 `Zone` 下转换。
3. **API 层**：新/改版接口 JSON 中时间字段 **MUST** 为数字（**Unix 秒**），与数据库存储一致；迁移期字符串须在 spec/runbook 中标为 DEPRECATED。
4. **Go 类型**：持久化层优先 `int64`；边界使用 `time.Unix(sec, 0)` / `.Unix()`，禁止将 `time.Local` 格式化为字符串再写入数字列。
5. **命名**：列名可沿用 `*_at`、`*_time` 等，注释统一为「Unix 秒 UTC」；禁止注释写「UTF8 日期」类易歧义表述。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 回填脚本误转时区导致全局偏移 | 回填前在 staging 对样本行做 `FROM_UNIXTIME(秒列)` 抽查；脚本内强制 UTC 解析源数据 |
| 旧客户端只认字符串 | 明确弃用期与版本门槛；runbook 定义切换读窗口 |
| gen 代码与手工 DDL 漂移 | 迁移 checklist：先 DDL 再 `gf gen dao`（或等价流程）并 CI 校验 |
| 亚秒事件丢失 | 本业务以秒为权威；若未来需要毫秒，须全链版本化升级单位 |

## Migration Plan

1. **盘点**：见 `inventory.md`；按服务列出库表列，标记当前类型与样本值。
2. **DDL**：新增 `*_sec BIGINT` 或 `MODIFY` 原列；旧列只读或双写由窗口决定（参考 `hack/ddl_timestamp_seconds_reference.sql` 注释模板）。
3. **回填**：`UPDATE` 将旧字符串解析为 UTC 秒后写入；校验 `COUNT(*)`、min/max 合理区间。
4. **应用**：切换读写至新列；`api/v1` 与 service 同步改为 `int64` JSON。
5. **回滚**：保留旧列至稳定期；读开关指回旧列。

## Open Questions

- 「仅日期无时刻」字段（如生日）是否改为 `DATE` 或仍用业务约定字符串：见 `appendix-exemptions.md`。

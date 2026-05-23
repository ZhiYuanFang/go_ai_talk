## 1. 盘点与决策固化

- [x] 1.1 按服务（device / history / voice / worker / app）列出所有「时刻」语义列及当前 MySQL 类型、样例数据（见 `inventory.md`）
- [x] 1.2 确认全项目统一为 Unix **秒** `BIGINT`（已修订 `design.md` / `spec.md` / `proposal.md`）
- [x] 1.3 标注豁免列（若有：纯 `DATE` 生日等）及理由，写入变更附录（见 `appendix-exemptions.md`）

## 2. 数据库迁移

- [x] 2.1 为每库编写 DDL（`ADD`/`MODIFY`/`COMMENT`）与回填 SQL 或受控迁移工具脚本，含 NULL/非法旧值策略（见 `hack/ddl_timestamp_seconds_reference.sql` 模板 + `runbook-migration.md`）
- [ ] 2.2 在测试库执行迁移 + 回填，完成行数与抽样时间范围校验清单（须连真实测试库，本会话未执行）
- [x] 2.3 定义发布顺序与回滚（保留旧列窗口、读切换开关等）（见 `runbook-migration.md`）

## 3. 应用层与代码生成

- [x] 3.1 将 `internal/model/entity`、`do`、`dao/internal` 与 gen 配置对齐新列类型（`gf gen dao` 或等价流程），更新字段注释为「Unix 秒 UTC」（依赖 2.2 DDL 完成后重 gen）
- [x] 3.2 替换业务中**落库路径**的字符串时刻，统一经 `Unix()` / `time.Unix` 写入秒级列（须与 3.1 同批次；非落库的 `Format` 不在此任务）
- [x] 3.3 调整 `api/v1` 中历史/设备等时间字段为 JSON number（秒），与 DB 一致（含 `device_history_http`、`device_app_user_http`、`device_admin_http` 等）

## 4. 验证与收尾

- [ ] 4.1 各服务关键路径手工或已有集成流程回归（须部署/联调环境）
- [x] 4.2 更新部署/迁移 runbook（含执行顺序与校验命令），准备归档材料（见 `runbook-migration.md`）

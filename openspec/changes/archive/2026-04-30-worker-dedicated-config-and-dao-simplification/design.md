## Context

当前项目已完成 `voice/device/history` 专属配置拆分，并将网关职责收敛为流量与策略层；但 `worker-service` 仍依赖主配置 `manifest/config/config.yaml`。同时，代码中仍保留了历史单体时期的多数据库组选择逻辑（`internal/dao/domain_db.go`）与部分 DAO 扩展层习惯，导致当前“服务=数据库”的边界规则在工程层面不够一致。

本次变更同时涉及配置体系、DAO 访问约束与运维文档治理，属于跨模块收敛改造。

## Goals / Non-Goals

**Goals:**
- 为 `worker-service` 提供专属配置，并在启动、compose、kustomize、Dockerfile 中统一切换。
- 删除主配置中的数据库项，确保其仅承载网关/公共配置。
- 移除 DAO 多数据库组分流机制，统一为服务进程内 `database.default` 访问模型。
- 评估 DAO 扩展文件的必要性并收敛为“最小可维护面”。
- 将 `dao-sync-by-domain.md` 与 `release-deploy-and-run.md` 迁移到统一文档目录并更新内容。
- 在 `openspec/project.md` 固化“服务-数据库一一对应、跨服务走 API”原则及文档更新要求。

**Non-Goals:**
- 不在本次变更中重构业务领域逻辑（voice/device/history 功能语义保持不变）。
- 不引入新的跨服务通信协议或消息中间件能力。
- 不重做历史发布体系，仅在现有部署资产上完成配置路径与说明更新。

## Decisions

1. **worker 配置独立化，主配置彻底去数据库**
   - 决策：新增 `manifest/config/config.worker-service.yaml`，并将 worker 运行默认读取该文件。
   - 原因：worker 当前承载 MQ 消费与 outbox relay，存在数据库访问；继续依赖主配置会让网关与 worker 的边界耦合。
   - 备选：
     - 维持 worker 使用主配置：短期改动小，但长期会持续模糊边界，且与专属配置策略冲突。

2. **DAO 访问模型改为“仅 default”**
   - 决策：移除 `internal/dao/domain_db.go` 多组选择回退逻辑，DAO 侧仅使用进程 `default` 数据库连接。
   - 原因：项目已转向“单服务单数据库”，多组分流入口会形成误导和隐性跨库风险。
   - 备选：
     - 保留多组逻辑以兼容：会保留历史复杂度，且与当前架构约束相冲突。

3. **DAO ext 文件按“必要性”保留，不做机械全删**
   - 决策：逐文件评估 `*_ext.go` 是否包含业务语义增强、读写库策略或查询聚合；无增量价值则合并/删除，有价值则保留并补充说明。
   - 原因：全删可能破坏既有 DAO 扩展语义；机械保留则会维持冗余。
   - 备选：
     - 全量删除 ext：风险高，难以控制回归。
     - 全量保留 ext：无法达成“简化到单服务形态”目标。

4. **运行文档集中管理**
   - 决策：创建新的运行文档目录（例如 `docs/runbooks/`），迁移 `dao-sync-by-domain.md` 与 `release-deploy-and-run.md`，并在 `project.md` 规定后续需求变更同步更新这两份文档。
   - 原因：当前运行文档分散，修改后易遗漏同步。
   - 备选：
     - 原地保留：路径不变但治理约束弱，仍易遗漏。

## Risks / Trade-offs

- **[Risk] worker 专属配置数据库若指向错误库，将导致 outbox relay 处理异常**  
  → **Mitigation**：在变更任务中增加 worker 启动验证与 outbox 冒烟检查（表存在、可读、可更新状态流转）。

- **[Risk] 移除 `domain_db.go` 后，个别历史调用可能隐式依赖 group fallback**  
  → **Mitigation**：全仓扫描 DAO 调用路径，统一改为 default 语义；编译 + 关键链路验证。

- **[Risk] DAO ext 收敛可能误删业务增强逻辑**  
  → **Mitigation**：先建立 ext 清单并分类（保留/合并/删除），每类给出依据；变更分批提交并做回归验证。

- **[Trade-off] 文档迁移会改变团队习惯路径**  
  → **Mitigation**：在原路径保留重定向说明（或索引链接），并在 `project.md` 明确唯一维护位置。

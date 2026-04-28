---
name: openspec
description: >-
  指导基于规格的变更提案、增量规格、校验与归档，配合 OpenSpec CLI 与本仓库目录结构。
  在用户提到 OpenSpec、变更提案、规格增量、openspec validate、规划功能、破坏性变更、
  架构调整，或要求创建/应用规格与变更时使用。
---

# OpenSpec（本仓库）

## 权威文档

编写或修改 OpenSpec 产物前，请先阅读：

1. **[openspec/AGENTS.md](../../../openspec/AGENTS.md)** — 完整工作流、CLI、增量格式、常见坑、排错。
2. **[openspec/project.md](../../../openspec/project.md)** — 本仓库约定。

若工作区规则（`AGENTS.md` 托管块）与本技能不一致，在运行 `openspec update` 后，以 **`openspec/AGENTS.md`** 为准。

## 何时走 OpenSpec

**创建变更**（提案 + 增量 + `tasks.md`，并用 `--strict` 校验）适用于：新增能力、破坏性 API/架构/安全模式变更、或改变行为的性能优化。

**可跳过提案**的情况：恢复规格所描述行为的缺陷修复、错别字与排版、非破坏性依赖升级、仅配置调整、或为既有行为补测试——除非需求模糊，走提案更安全。

## 助手工作流（精简）

### 编写或实现之前

- 运行 `openspec list`、`openspec list --specs`（或 `openspec spec list --long`）了解上下文。
- 阅读相关 `openspec/specs/<capability>/spec.md`；检查 `openspec/changes/` 是否已有重叠变更。
- 按需使用 `openspec show <spec-id> --type spec` / `openspec show <change-id> --json --deltas-only`。

### 创建变更

1. 选取唯一的动词开头 **change-id**（kebab-case：`add-…`、`update-…`、`remove-…`、`refactor-…`）。
2. 在 `openspec/changes/<change-id>/` 下搭建：`proposal.md`、`tasks.md`、可选 `design.md`（是否撰写见 AGENTS.md 条件），以及 `specs/<capability>/spec.md` 增量。
3. 增量使用 `## ADDED|MODIFIED|REMOVED|RENAMED Requirements`；每条 Requirement 至少包含一个 `#### Scenario:` 块（标题格式须与 AGENTS.md 一致，四个 `#`）。
4. 分享前运行 **`openspec validate <change-id> --strict`** 并修复全部问题。
5. 提案**未经评审通过前不要开始实现**。

### 实现已批准的变更

按顺序阅读 `proposal.md` → `design.md`（若有）→ `tasks.md`；完成任务后，如实勾选 `tasks.md` 中的复选框。

### 上线之后

按 AGENTS.md 归档（含适时使用 `openspec archive <change-id> …`）；归档后再跑 `openspec validate --strict`。

## CLI 速查

```bash
openspec list                  # 进行中的变更
openspec list --specs          # 规格列表
openspec spec list --long      # 规格（详细）
openspec show <item>           # 查看变更或规格
openspec validate <id> --strict
openspec archive <change-id> [--skip-specs] [--yes]
openspec update [path]         # 刷新说明文件
```

## 常见坑（细节见 AGENTS.md）

- **MODIFIED 需求**：从 `openspec/specs/...` 粘贴**完整**更新后的需求块；增量若不全，归档时会丢细节。
- **Scenario**：仅使用 `#### Scenario: 标题` —— 不要用列表项、加粗或 `###` 等变体。
- **检索**：优先用 `openspec` 子命令，并对 `openspec/specs`、`openspec/changes` 使用 `rg`，少凭感觉猜。

## 延伸阅读

- 增量示例与多能力布局：**openspec/AGENTS.md** 中「Creating Change Proposals」「Multi-Capability Example」「Troubleshooting」等章节。

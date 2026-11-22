---
name: openspec
description: >-
  指导基于规格的变更提案、增量规格、校验与归档，配合 OpenSpec CLI 与本仓库目录结构。
  在用户提到 OpenSpec、变更提案、规格增量、openspec validate、规划功能、破坏性变更、
  架构调整，或要求创建/应用规格与变更时使用。
---

# OpenSpec（本仓库）

## 权威文档

1. **[openspec/project.md](../../openspec/project.md)** — 唯一项目上下文（布局、归档、工程指针）。
2. **`openspec/specs/<domain>/spec.md`** — 域行为真相（`openspec list --specs`）。

## 何时走 OpenSpec

**创建变更**适用于：新增能力、破坏性 API/架构/安全模式变更、或改变行为的性能优化。

**可跳过提案**：恢复规格所描述行为的缺陷修复、错别字、非破坏性依赖升级、仅配置调整——除非需求模糊。

## 助手工作流

### 编写或实现之前

- 阅读 `openspec/project.md` 与相关域规格。
- 运行 `openspec list`、`openspec list --specs`。
- 检查 `openspec/changes/` 是否已有重叠变更。

### 创建变更

1. 选取唯一的动词开头 **change-id**（kebab-case）。
2. `openspec new change "<id>"`，按 schema 写 `proposal` → `specs` → `design` → `tasks`。
3. 每条 Requirement 至少一个 `#### Scenario:`；正文含 **SHALL** 或 **MUST**。
4. 分享前 `openspec validate <change-id> --strict`。

### 实现已批准的变更

阅读 `proposal` → `design` → `tasks`；对照相关域规格；完成后勾选 `tasks.md`。

### 归档（本仓库定制）

官方语义：把 delta 合并进 `openspec/specs/<domain>/spec.md`。

本仓库（单人 OPC）额外约定：

1. 若实现阶段已直接写妥主规格，可用 `openspec archive <id> -y --skip-specs`（或等价跳过合并）。
2. 否则 `openspec archive <id> -y` 合并 delta。
3. **删除** `openspec/changes/<id>/`（以及若 CLI 生成了 `changes/archive/YYYY-MM-DD-<id>/`，一并删除）。
4. **默认不保留** archive 历史目录。

## CLI 速查

```bash
openspec list
openspec list --specs
openspec status --change "<name>" --json
openspec validate <id> --strict
openspec archive <id> -y
openspec archive <id> -y --skip-specs
```

## 延伸阅读

- 命令：`/opsx-propose`、`/opsx-apply`、`/opsx-archive`、`/opsx-explore`
- 技能：`.cursor/skills/openspec-propose/`、`openspec-apply-change/`、`openspec-archive-change/`

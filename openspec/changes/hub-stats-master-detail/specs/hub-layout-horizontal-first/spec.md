## ADDED Requirements

### Requirement: Hub 运维页优先左右分栏

`openspec/project.md` 与 `AGENTS.md` MUST 增加 Hub 运维页布局约定：新增或改造 Hub 静态运维页时，主列表与详情/侧栏内容在宽屏 MUST 优先左右利用空间；禁止将主从内容默认整页纵向堆叠导致长时间下滚（窄屏允许回退为上下布局）。

#### Scenario: 规范文档可检索

- **WHEN** 评审或 AI 查阅全局约定
- **THEN** 可在 `openspec/project.md` 与 `AGENTS.md` 中找到「Hub 运维页布局」相关强制条款，并包含宽屏左右、窄屏可堆叠、后续 proposal/design 须说明空间利用方式等要点

#### Scenario: 后续变更受约束

- **WHEN** 后续 OpenSpec 变更新增或改造 Hub 运维页且存在「列表 + 详情」交互
- **THEN** proposal/design MUST 说明宽屏左右分栏方案（或书面说明为何不适用），评审 MUST 检查是否违反该约定

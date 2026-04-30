## Why

当前变更文档存在中英文混用与表达风格不统一的问题，增加了团队沟通成本与评审理解偏差。需要将 OpenSpec 产物统一为中文，以保障后续提案、设计、规格与任务的一致性和可维护性。

## What Changes

- 将 OpenSpec 变更文档语言规范统一为中文，覆盖 proposal、design、specs、tasks 全部工件。
- 建立文档语言校验约束：新增或更新变更时，必须使用中文撰写说明性内容。
- 明确术语书写规则：保留必要英文技术标识（环境变量、路径、接口、协议关键字），其余内容默认中文。
- **BREAKING**：不再接受以英文为主的 OpenSpec 工件作为可实施输入。

## Capabilities

### New Capabilities
- `chinese-documentation-governance`: 规范 OpenSpec 工件默认语言为中文，并提供一致的术语与写作约束。
- `documentation-language-compliance`: 定义变更文档在提案、设计、规格、任务阶段的语言合规要求。

### Modified Capabilities
- None.

## Impact

- 影响范围：`openspec/changes/*` 下的新建与增量工件内容。
- 影响流程：提案创建、设计评审、任务分解与实施前检查均需遵循中文化规范。
- 对代码运行时无直接影响，主要提升文档协作质量与交付一致性。

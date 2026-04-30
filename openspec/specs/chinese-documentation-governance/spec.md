# chinese-documentation-governance Specification

## Purpose
TBD - created by archiving change enforce-chinese-documentation. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec 工件默认使用中文
系统在创建或更新 OpenSpec 变更工件时，说明性文本 SHALL 使用中文撰写，包括 proposal、design、specs、tasks。

#### Scenario: 创建新变更工件
- **WHEN** 用户通过 OpenSpec 工作流生成新的 proposal/design/specs/tasks
- **THEN** 生成内容中的说明性文本 SHALL 为中文

### Requirement: 必要技术标识允许保留英文
系统在文档中文化过程中 SHALL 允许保留必要英文技术标识，包括环境变量、路径、接口、协议和代码符号。

#### Scenario: 文档包含技术标识
- **WHEN** 文档中出现环境变量名、API 路径或代码符号
- **THEN** 这些标识 SHALL 保持英文原文，不做强制翻译


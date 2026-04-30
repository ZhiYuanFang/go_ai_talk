## ADDED Requirements

### Requirement: 变更文档需要通过语言合规检查
系统在变更进入实施阶段前 SHALL 完成文档语言合规检查，若说明性文本以英文为主则不得作为可实施输入。

#### Scenario: 变更进入 apply 前校验
- **WHEN** 变更已生成 proposal、design、specs、tasks 并准备进入实施阶段
- **THEN** 文档语言合规检查 SHALL 确认说明性文本为中文，否则应阻止进入实施

### Requirement: 语言规则适用于增量更新
系统在对已有变更进行增量更新时 SHALL 继续遵循中文文档规则，不因历史内容存在英文而豁免。

#### Scenario: 更新已有变更工件
- **WHEN** 用户对现有变更工件进行追加或修订
- **THEN** 新增与修改内容 SHALL 使用中文说明并保持术语一致

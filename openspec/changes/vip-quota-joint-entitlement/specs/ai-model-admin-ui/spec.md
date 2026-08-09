## ADDED Requirements

### Requirement: AI 模型与并发页展示 careAlert 与 free

`/device/admin/ai-model-admin.html` MUST 在 Voice 区域展示 **三条** lane：`voiceUnderstanding`、`clinic`、`careAlert`；对上述三条及 UCG `polish` MUST 提供 freeProvider/freeModel 编辑控件（允许清空）。Sim 四条 MUST NOT 展示 free 编辑。保存时 MUST 将 free 字段一并提交至对应域 Admin API。页面仍为运维型，MUST NOT 计入 App usage 统计。

#### Scenario: 页面加载三条 Voice lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 careAlert 块及各业务 lane 的 free 字段

#### Scenario: 清空 free 可保存

- **WHEN** 管理员清空某业务 lane 的 free 并保存成功
- **THEN** 后续非 premium 路径 MUST 按 omit/Python 自选语义工作

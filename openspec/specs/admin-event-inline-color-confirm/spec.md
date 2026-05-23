# admin-event-inline-color-confirm Specification

## Purpose
TBD - created by archiving change admin-event-inline-color-confirm. Update Purpose after archive.
## Requirements
### Requirement: 行内色调编辑须确认后保存

事件管理页在行内修改事件色调时，SHALL 提供明确的用户确认步骤；系统 SHALL NOT 在取色器 `change` 事件发生时自动调用更新接口。

#### Scenario: 打开色调编辑浮层

- **WHEN** 已登录管理员点击某行「色调」展示区域
- **THEN** 页面 SHALL 显示包含取色控件及 **确定**、**取消** 控件的编辑浮层
- **AND** 取色器 SHALL 初始化为该行当前 `color`（合法时）或默认色

#### Scenario: 点击确定后提交

- **WHEN** 用户在浮层中调整颜色并点击 **确定**
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update` 并携带该行完整字段与新 `color`
- **AND** 成功后 SHALL 关闭浮层并刷新事件列表

#### Scenario: 点击取消不提交

- **WHEN** 用户在浮层中点击 **取消** 或等价取消操作
- **THEN** 系统 SHALL NOT 调用更新接口
- **AND** 浮层 SHALL 关闭且列表数据保持不变

#### Scenario: 提交中防重复

- **WHEN** 色调更新请求正在进行
- **THEN** 浮层内 **确定** 按钮 SHALL 处于不可用或加载状态直至请求结束

### Requirement: 弹窗编辑与其它行内能力不受影响

弹窗「编辑事件」中的色调与提交行为 SHALL 保持可用；行内 Logo 点击上传流程 SHALL 不因本变更而不可用。

#### Scenario: 弹窗编辑仍可用

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 仍打开含色调字段的编辑弹窗并按原逻辑提交


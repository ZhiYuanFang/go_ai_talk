# device-admin-event-logo-color-ui Specification

## Purpose
TBD - created by archiving change device-admin-event-logo-color-ui. Update Purpose after archive.
## Requirements
### Requirement: 事件管理列表展示 Logo 与色调

设备管理页（`admin.html` 或等价路由）在登录并加载事件列表后，SHALL 在表格中展示 **Logo** 与 **色调** 列；每行 SHALL 根据 `GET /device/admin/api/event/list` 返回的 `logo`、`color` 渲染预览。

#### Scenario: 列表含 logo 与 color 字段时展示预览

- **WHEN** 事件列表项包含 `logo` 路径与有效 `color` 色值
- **THEN** 页面 SHALL 在 Logo 列显示可识别的缩略图
- **AND** 色调列 SHALL 显示与 `color` 一致的色块及可读色值文本

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

### Requirement: 管理页 Logo 预览使用同源 URL

管理页用于 `<img src>` 的 logo 地址 SHALL 使用**当前页面所在 origin** 与库中 path 拼接，SHALL NOT 默认拼接至 App 网关（:9702）基址。

#### Scenario: path-only logo 同源加载

- **WHEN** `logo` 为 `/ai_talk_images/event_1.png` 且管理页为 `https://example.com:9701/device/admin`
- **THEN** 图片请求 URL SHALL 为 `https://example.com:9701/ai_talk_images/event_1.png`

#### Scenario: 历史绝对 URL 仍可显示

- **WHEN** `logo` 已是 `http://` 或 `https://` 开头的绝对 URL
- **THEN** 页面 MAY 直接使用该 URL 显示（兼容迁移数据）

### Requirement: 主网关提供同源事件图片访问

gateway-service（管理页常用入口，如 :9701）SHALL 注册 `GET /ai_talk_images/*`（及 HEAD），将请求反代或等价转发至 device-service 的同名静态读能力，使管理页同源 URL 可成功返回图片。

#### Scenario: 经主网关读取已上传 logo

- **WHEN** 客户端请求 `GET https://<gateway-host>/ai_talk_images/<安全文件名>` 且 device-service 上文件存在
- **THEN** gateway-service SHALL 返回对应图片内容且 SHALL NOT 要求 App 网关 Bearer

### Requirement: 点击色调即可更新 color

管理页 SHALL 允许用户通过点击列表行中的色调展示区域修改该事件的 `color`，并在成功后刷新列表。

#### Scenario: 点击色块修改 color

- **WHEN** 用户点击某行色调区域并选择新色值后确认提交
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update`（multipart）并携带该行 `id`、`name`、`needQuantity`、`extraNames` 及新 `color`
- **AND** 未选择新 logo 文件时 SHALL NOT 清除原有 `logo`

#### Scenario: 更新成功后列表反映新色值

- **WHEN** 更新接口返回成功
- **THEN** 页面 SHALL 刷新事件列表且该行色调展示与新 `color` 一致

### Requirement: 点击 Logo 即可更新 logo 文件

管理页 SHALL 允许用户通过点击列表行中的 Logo 缩略图或占位触发文件选择，上传新图并更新该事件。

#### Scenario: 点击 Logo 上传新图

- **WHEN** 用户点击 Logo 区域并选择合法图片文件（如 png/jpeg/webp）
- **THEN** 客户端 SHALL 以 multipart 调用 `POST /device/admin/api/event/update`，包含 `logo` 文件及该行完整文本字段
- **AND** 成功后服务端 `event.logo` SHALL 更新为 path-only 新路径

#### Scenario: 更新成功后列表展示新缩略图

- **WHEN** logo 更新成功且列表重新加载
- **THEN** 该行 Logo 列 SHALL 使用同源 URL 展示新图

### Requirement: 行内编辑与弹窗编辑并存

名称、事件扩展、是否需要计数等字段 SHALL 仍可通过既有「编辑」弹窗修改；行内交互仅负责 **logo** 与 **color**，SHALL NOT 要求用户仅为改色/改图打开完整弹窗。

#### Scenario: 编辑按钮仍打开完整表单

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 打开包含名称与其它字段的编辑弹窗（行为与变更前一致）


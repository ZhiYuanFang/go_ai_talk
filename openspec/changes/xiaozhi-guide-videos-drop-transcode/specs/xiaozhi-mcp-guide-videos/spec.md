## ADDED Requirements

### Requirement: cash SHALL 持久化小智引导视频双固定槽

cash-service MUST 维护固定两槽引导视频元数据，槽标识 MUST 为 `promo`（能力宣传）与 `tutorial`（绑定教学）。每槽 MUST 含：可编辑 `title`、可选视频 `object_key`（库内仅存 OSS objectKey，MUST NOT 存完整 CDN URL 作为权威源）、`updated_at`。空 `object_key` MUST 表示该槽未配置。系统 MUST 在 schema 初始化时种子两行（可带默认中文标题），Hub 可随后改标题。

#### Scenario: 种子两槽存在

- **WHEN** cash schema 初始化完成
- **THEN** `promo` 与 `tutorial` 两行 MUST 存在且均可被 Admin 读取

#### Scenario: 清空 object_key 表示未配置

- **WHEN** 某槽 `object_key` 为空
- **THEN** App 列表 MUST NOT 包含该槽

### Requirement: Hub Admin SHALL 管理引导视频标题与上传

系统 MUST 提供经 cash Admin 鉴权的接口，支持：列出两槽（含 CDN 预览 URL）、更新某槽标题、multipart 上传/覆盖某槽视频、清空某槽视频（清除 `object_key`）。上传 MUST 经 `clients/ucg` 写入 OSS（建议前缀 `xiaozhi-guide/`），MUST 使用与 UCG App 视频上传相同的 **v1 验真与体积上限**，MUST NOT 走服务端转码，MUST NOT 落 gateway 本机盘。Hub UI MUST 挂在现有「小智 MCP 开通」页（区 D），MUST NOT 要求新建独立导航模块（可复用既有开通页入口）。

#### Scenario: 运维上传合规视频

- **WHEN** 运维对 `tutorial` 槽上传通过 v1 验真的 mp4
- **THEN** 该槽 `object_key` MUST 更新，Admin GET MUST 返回可访问的 `videoCdnUrl`

#### Scenario: 不合规视频拒绝

- **WHEN** 运维上传未通过 v1 验真的文件
- **THEN** API MUST 返回错误且 MUST NOT 更新该槽已有 `object_key`（若曾配置）

#### Scenario: 改标题不强制已有视频

- **WHEN** 运维仅更新 `promo` 标题且该槽尚无视频
- **THEN** 标题 MUST 保存成功

### Requirement: App SHALL 返回已配置引导视频的 CDN 列表

系统 MUST 提供 `GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）。响应 `list` MUST 仅含 `object_key` 非空的槽，每项至少含 `slot`、`title`、`videoCdnUrl`；MAY 含 `coverCdnUrl`（视频首帧缩略图 CDN）。该接口 MUST 计入 gateway-app usage 统计（MUST NOT 加入维护型 skip 列表）。

#### Scenario: 仅一槽有视频

- **WHEN** 仅 `promo` 已配置视频
- **THEN** `list` 长度 MUST 为 1 且含 `slot=promo`

#### Scenario: 未登录拒绝

- **WHEN** 请求无有效 App Bearer
- **THEN** MUST 拒绝（MUST NOT 匿名放行）

### Requirement: Flutter 小智页 SHALL 按清单展示或隐藏入口

Flutter 客户端（`flutter_ai_talk`）在小智音箱相关页面 MUST 请求引导视频清单。当 `list` 为空时 MUST **完全不展示**引导视频入口。当存在一项或两项时 MUST 按返回的 `title` 展示可点入口，点击后 MUST 使用返回的 `videoCdnUrl` 播放。MUST NOT 在双槽皆空时展示占位「暂无视频」入口块（与「完全不露」一致）。

#### Scenario: 双空隐藏

- **WHEN** App 返回 `list` 为空数组
- **THEN** 小智页 MUST 不渲染引导视频入口 UI

#### Scenario: 有视频可点播

- **WHEN** `list` 含 `tutorial` 项
- **THEN** 用户点击对应入口后客户端 MUST 能打开该 `videoCdnUrl` 进行播放

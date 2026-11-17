## MODIFIED Requirements

### Requirement: 官网展示 Android 下载二维码与 iOS 指引

官网 SHALL 提供独立的应用下载区块。Android 下载区 MUST 基于数据库中的最新下载链接生成二维码并展示可点击下载入口：当最新行为 OSS objectKey 时，聚合接口 MUST 返回 CDN 绝对 URL；当最新行为历史 path-only 时，MUST 仍拼出 gateway 可访问的绝对地址。iOS 下载区 MUST 提示用户前往 App Store 搜索“胖宝”下载。

#### Scenario: Android 存在可下载版本（OSS）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `apk/` objectKey
- **THEN** 官网聚合接口 SHALL 返回 CDN 绝对 Android 下载地址，页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 存在可下载版本（历史 path）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `/device/app/apk/` 路径且本地文件可访问
- **THEN** 官网聚合接口 SHALL 返回官网可直接使用的绝对下载地址（基于 `publicBaseUrl` 或请求 Host），页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 暂无可下载版本

- **WHEN** 版本表没有可用的 Android 下载记录
- **THEN** 页面 SHALL 不展示失效二维码，并 SHALL 展示明确的“Android 下载暂未开放”或等价提示

#### Scenario: iOS 下载说明固定展示

- **WHEN** 用户查看官网下载区
- **THEN** 页面 SHALL 展示“前往 App Store 搜索‘胖宝’下载”的文案，而不要求数据库提供 iOS 下载链接

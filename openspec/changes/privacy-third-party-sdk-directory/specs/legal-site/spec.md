## ADDED Requirements

### Requirement: 隐私政策 MUST 披露第三方 SDK 目录

html` MUST 提供独立的「第三方 SDK 目录」（或等价标题）表格式清单，至少覆盖现网 Android 客户端会集成的下列类别（表述可润色）：

- 微信 OpenSDK（登录）
- 厂商推送：华为推送、小米推送、荣耀推送、OPPO 推送、VIVO 推送（经 china_push 等集成； 可说明按设备厂商通道启用）
- 支付宝 SDK（Android 付费开通主路径）
- 云端语音识别服务（MUST 不点名具体云厂商； 用途为语音转文字）
- 本地语音唤醒组件（如 sherpa_onnx：端侧处理）
- 设备信息相关组件（与设备型号披露一致）
- 扫码/相册/录音等多媒体相关组件（可合并一行，与权限用途一致）

每行 MUST 至少含：SDK/服务名称、提供方、用途、可能涉及的个人信息种类； MUST 提供对方隐私政策链接或可访问的说明入口（云端语音识别不点名时 MAY 省略具体厂商链，但 MUST 在用途中说明经具备相应资质的云端语音识别服务处理）。 目录 MUST NOT 将 `geolocator` 或定位 SDK.

#### Scenario: 用户查阅 SDK 目录

- **WHEN** 用户打开隐私政策并阅读第三方 SDK 目录
- **THEN** 页面 SHALL 以表格或等价结构化列表展示上述 SDK/服务，且 SHALL 含微信、支付宝、以及华为/小米/荣耀/OPPO/VIVO 推送相关条目

#### Scenario: 目录不含定位与应用内下载

- **WHEN** 用户阅读第三方 SDK 目录与全文
- **THEN** 文案 MUST NOT 将定位 SDK 或应用内 APK 下载/自更新描述为现行收集或现行功能

## Context

- Explore 已定：SDK 目录含 china_push 连带的 OPPO/VIVO；geolocator 发布包剔除；支付以支付宝为主；隐私不写应用内下载（更新逻辑另改）。
- 隐私政策已去 AIGC/点名 DashScope；本变更只补 SDK 清单与推送厂商对齐。

## Goals / Non-Goals

**Goals:**

- 隐私政策提供可审核的第三方 SDK 表格式目录。
- 推送：华为、小米、荣耀、OPPO、VIVO 均列出。
- 发布包不再依赖 geolocator。

**Non-Goals:**

- 不改 china_push 集成范围、不在本变更改应用内更新实现。
- 不恢复定位权限或 UCG 广场距离能力。
- 不在 SDK 表罗列纯 UI 依赖（Riverpod 等）。

## Decisions

1. **SDK 表字段**：名称 | 提供方 | 用途 | 可能涉及的信息 | 对方隐私政策（外链）。用 HTML `<table>` 适配窄屏横向滚动。
2. **必列行**：微信 OpenSDK、华为/小米/荣耀/OPPO/VIVO 推送、支付宝、云端语音识别（不点名）、sherpa 本地唤醒、设备信息类、扫码/相册/录音相关（可合并一行）。Apple 登录注「仅 iOS」。
3. **支付**：Android 主路径写支付宝；不强调 Google Play IAP。
4. **geolocator**：`pubspec` 移除；`ucg_location.dart` 改为无 Geolocator 的桩（恒失败/未启用），避免大范围删 UCG 文件；Manifest 定位权限保持注释关闭。
5. **不写**：REQUEST_INSTALL_PACKAGES / 应用内下载。

## Risks / Trade-offs

- [Risk] 厂商隐私链 URL 变更 → 用各开放平台稳定「个人信息处理规则」入口页。  
- [Risk] UCG 定位桩导致若误开 UCG 入口无定位 → 可接受（入口关闭）。  
- [Risk] OPPO/VIVO 未配密钥仍列 SDK → 符合「包内集成」披露，文案注明按机型通道启用。

## Migration Plan

1. 更新隐私 HTML 并部署 gateway-app-server。  
2. Flutter 剔 geolocator 后打华为渠道包。  
3. 回滚：还原 HTML 与 pubspec。

## Open Questions

（无）

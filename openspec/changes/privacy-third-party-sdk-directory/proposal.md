## Why

华为等应用市场上架要求隐私政策列出客户端第三方 SDK 目录；现网隐私仅概括「云端语音识别 / 推送 / 支付」，未按 SDK 披露微信、china_push 各厂商推送、支付宝等。同时发布包仍含未启用的定位库 `geolocator`，易被扫包质疑。

## What Changes

- **隐私政策**（`privacy-policy.html`）：新增「第三方 SDK 目录」表（名称、提供方、用途、信息种类、隐私政策链接）；推送厂商含华为/小米/荣耀/**OPPO/VIVO**（随 `china_push` 集成）；Android 付费以**支付宝**为主表述；**不写**应用内 APK 下载/自更新；云端语音识别继续不点名；更新修订日期。
- **Flutter 发布包**（`flutter_ai_talk`）：从依赖与调用路径**剔除 `geolocator`**，保证 UCG 相关代码在无定位插件下可编译（入口关闭时亦不打进定位库）。
- **规格**：修订 `privacy-policy-company`（SDK 目录要求）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `privacy-policy-company`：MUST 披露第三方 SDK 目录；推送厂商与支付宝/微信等对齐现网；MUST NOT 将定位 SDK 或应用内下载写成现行能力。

## Impact

- `resource/public/privacy-policy.html`（`gateway-app-server` 静态页）。
- 邻仓 `flutter_ai_talk/app`：`pubspec.yaml` 与 `ucg_location` 等引用。
- **部署**：`gateway-app-server`；Flutter 需重新出 Android 包。不必改服务端 `.env`。

## Why

现网隐私政策仍按「全量能力面」披露：微信头像/昵称、VIP、UCG 社区与位置、值得留意/成长轨迹/诊疗等。客户端已关闭 UCG 入口（`kUcgHomePagerEnabled=false` 等），VIP/值得留意/成长轨迹等已退役，微信登录仅持久化 `unionid`。文案与真实产品面不一致，需一次性对齐，避免过度或错误披露。

## What Changes

- 修订 `resource/public/privacy-policy.html`（路径不变），更新修订日期。
- **运营主体**：保留公司中英文全称与客服邮箱；**删除**注册地址相关表述。
- **微信**：改为仅说明用微信登录获取账号标识（unionid）；**明确不收集**微信头像、昵称。
- **商业化**：精简披露——用户付费或邀请开通部分功能时处理订单与权益状态；**不写 VIP**。
- **设备型号**：新增披露——用于崩溃排查与运维统计，非广告跟踪。
- **隐藏/删除**客户端已关闭的 UCG 相关收集项与权限（社区资料、互动、私信、媒体、属地、浏览/发帖位置、位置权限等）。
- **权限**：相机/相册改为扫码绑定、相册识码、成长记录配图、宝宝头像等现网用途；麦克风保留喂养/预测语音；通知去掉「值得留意」。
- **第三方 AI / 审核**：删除 UGC 润笔（或仅保留已下线一句）、诊疗、值得留意、成长轨迹、社区内容安全 Green；保留横屏喂养语音 AI 的多供应商可配置表述。
- 同步清洗「不收集」「第三方服务」「注销」等章节中的社区/位置残留，保持全文一致。
- **MODIFIED** OpenSpec `privacy-policy-company`，使其不再强制注册地址与过时能力清单。
- Flutter 仓**无需**改加载 URL（仍打开 `/privacy-policy.html`）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `privacy-policy-company`：主体披露去掉注册地址强制；收集/权限/AI/第三方范围对齐现网客户端表面（无 UCG 入口、无 VIP、含设备型号与精简商业化）。

## Impact

- **静态页**：`resource/public/privacy-policy.html`（经 `gateway-app-server` 对外）。
- **规格**：基线 `openspec/specs/v3.2.0` 中 `privacy-policy-company` 需求变更。
- **不必改**：Flutter App 代码路径；`user-agreement.html`（本变更不改用户协议，除非后续单独立项）。
- **部署**：更新静态资源后需重部署或刷新 `gateway-app-server` 所托管 public 资源。

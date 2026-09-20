## Why

运营客服邮箱由个人域 `tou_zy@foxmail.com` 更换为公司域 `pangbao@cuplay.top`。官网、隐私政策与用户协议当前仍展示旧地址，需与现网联系渠道一致，并同步更新 OpenSpec 基线中写死的邮箱文案。

## What Changes

- 将 `resource/public/pangbao-home.html`（官网）页脚客服邮箱改为 `pangbao@cuplay.top`（含 `mailto:`）。
- 将 `resource/public/privacy-policy.html`、`resource/public/user-agreement.html` 中全部客服/联系邮箱改为 `pangbao@cuplay.top`。
- 更新隐私政策与用户协议的**修订日期**（生效日期可保持不变）；路径仍为 `/privacy-policy.html`、`/user-agreement.html`。
- 同步修改规格：`company-site-legal`、`privacy-policy-company`、`user-agreement-company` 中 MUST 邮箱从 `tou_zy@foxmail.com` 改为 `pangbao@cuplay.top`。
- **无 BREAKING**：对外 URL、API、客户端加载路径不变。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `company-site-legal`：官网页脚客服邮箱改为 `pangbao@cuplay.top`。
- `privacy-policy-company`：隐私政策联系邮箱改为 `pangbao@cuplay.top`，并更新修订日期要求。
- `user-agreement-company`：用户协议客服邮箱改为 `pangbao@cuplay.top`，并更新修订日期要求。

## Impact

- **静态页**：`resource/public/pangbao-home.html`、`privacy-policy.html`、`user-agreement.html`（共约 5 处 mailto/展示文案）。
- **规格**：上述三个 capability 的 Requirement 文本。
- **发布**：经 gateway-app 托管的官网与合规页一并上线即可；不新增 App HTTP 路由、不改 usage 统计、不加 Redis、无背景循环。
- **运维前提**：`pangbao@cuplay.top` 须已能收信（本变更不负责邮箱系统开通）。

## Context

- 合规文档由 gateway-app 静态托管：`gateway_app_register.go` 绑定 `/privacy-policy.html` 与 `/user-agreement.html`。
- Apple 登录后端（`apple-sign-in-api`）仅将校验后的 JWT `sub` 写入 `wx.apple_sub`；**不**持久化 Apple 邮箱、姓名或 `identityToken` 原文（见 `apple_auth.go`、`design.md`）。
- 用户选定 **方案 A（最小）**：只披露 Apple，不顺带改写用户名/设备号登录段落。

## Goals / Non-Goals

**Goals:**

- 隐私政策如实描述 Apple 登录所收集/存储的数据边界，与实现一致。
- 用户协议说明 iOS 可使用 Apple 登录建立账户。
- 两份文档生效日期更新为修订日（2026-06-06）。

**Non-Goals:**

- 不新增章节描述用户名、设备号登录。
- 不新增「第三方共享」长章节或账号绑定/不可合并细则（留待方案 B/C）。
- 不修改 App Privacy 标签（Flutter 客户端侧）、不新增测试文件。

## Decisions

1. **隐私政策 §1 增补独立列表项**（而非重写微信条目），降低 diff 与误伤风险。
2. **措辞对齐后端**：写「Apple 提供的匿名用户标识符」，明确「不存储 Apple 邮箱、姓名或登录凭证原文」。
3. **用户协议 §2 增补一条** Apple 登录说明，保留原有微信两条不变；账号安全第二条泛化为「第三方账号」仍属方案 B，本变更仅加 Apple 专条。
4. **生效日期**：统一为 `2026-06-06`（与修订日一致）。

## Risks / Trade-offs

- 文档仍不完整覆盖用户名/设备号登录 → 已知取舍，后续可单独变更补齐。
- 未写 Apple 隐私政策外链 → 方案 A 可接受；若审核要求可后续追加。

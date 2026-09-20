## Context

客服联系邮箱从 `tou_zy@foxmail.com` 更换为 `pangbao@cuplay.top`。展示面为 gateway-app 静态资源：官网 `pangbao-home.html`、隐私政策、用户协议。v3.0.3 基线在 `company-site-legal`、`privacy-policy-company`、`user-agreement-company` 中写死了旧邮箱。产品要求官网与合规页一并上线。

## Goals / Non-Goals

**Goals:**

- 三页 HTML 全部可见邮箱与 `mailto:` 统一为 `pangbao@cuplay.top`。
- 隐私政策、用户协议更新「修订日期」（生效日期保持不变）。
- 三份 capability delta 与实现一致。

**Non-Goals:**

- 不改合规文档 URL；不改 App 客户端。
- 不开通/配置邮箱服务本身。
- 不改公司名、ICP、其它合规条款正文（除邮箱与修订日期外）。

## Decisions

### D1：官网与合规页同变更同发

- **选择**：同一 change / 同一发布，避免官网新邮箱、App 合规页仍显示旧地址。
- **备选**：只改合规页 — 已否决（用户明确「官网一并上线」）。

### D2：只改修订日期，不改生效日期

- **选择**：联系方式变更属修订；`privacy-policy` / `user-agreement` 的「修订日期」更新为发布日（实现时取当天日期，如 2026-09-20）。
- **备选**：同时改生效日期 — 过重，非协议整体重订。

### D3：全文替换旧邮箱字符串

- **选择**：对三份 HTML 做 `tou_zy@foxmail.com` → `pangbao@cuplay.top` 全量替换（含 mailto），并用 `rg` 验收仓库内无残留。
- **备选**：抽公共配置注入 — 静态 HTML 无模板引擎，过设计。

## Risks / Trade-offs

- **[邮箱未开通]** → 上线前运维确认 `pangbao@cuplay.top` 可收信；本变更无法技术校验。
- **[CDN/缓存旧页]** → 静态资源随 gateway-app 镜像发布；若有边缘缓存，按既有发版清缓存流程。
- **[规格漏改]** → tasks 显式勾选三 capability；归档时合并进版本基线。

## Migration Plan

1. 合并后发布 gateway-app（或承载 `resource/public` 的进程）。
2. 抽检：`/` 或官网入口页脚、`/privacy-policy.html`、`/user-agreement.html` 邮箱均为新地址。
3. 回滚：回退静态页与规格中的邮箱字符串即可。

## Open Questions

- （无阻塞）邮箱收信侧由运维/业务确认，不阻塞本 change 文档与实现。

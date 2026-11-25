# 商户台 / 多宝宝 — 部署与冒烟

## 需重部署进程

- `device-service`（merchant 表、成员表、商户 API）
- `cash-service`（merchant_entitlement / product / order、支付宝分流）
- `history-service`（create 写闸）
- `gateway-app-server`（商户 JWT、SSE、`/merchant` SPA、反代）

## 入口

| 面 | URL |
|----|-----|
| 商户 SPA | `/merchant`、`/merchant/index.html`（勿再单独注册 `/merchant/`，与前者同路由） |
| 聚合注册/登录 | `POST /device/app/api/merchant_register` · `merchant_login` |
| Hub 运维 | `/device/admin/cash-merchant-admin.html` |
| 商户 SSE | `GET /device/merchant/sse/babies`（Bearer 商户 JWT） |

## 环境变量

- 既有：`GATEWAY_APP_JWT_SECRET`、`DEVICE_SERVICE_URL`、`CASH_SERVICE_URL`、`DEVICE_GATEWAY_INTERNAL_SECRET`
- 支付宝（生产）：`CASH_ALIPAY_*`；开发模拟履约：`CASH_PAYMENT_DEV_BYPASS=1` 后调  
  `POST /cash/merchant/api/subscribe/simulate-paid` `{ "orderNo": "..." }`

## 冒烟清单

1. 打开 `/merchant` 注册店名密码 → 拿到 accessToken  
2. 创建 2 个宝宝；第 3 个应失败（未授期）  
3. Hub 对 merchantId 设绝对到期日 → 再创建第 3 个成功  
4. 创建 staff、分发宝宝；App 用户名登录 staff  
5. 欠费后 App 新增 history 失败；end/update 仍可  
6. 商户台 SSE 在 history create 后局部刷新卡片文案  
7. `subscribe` 建单 →（DEV）`simulate-paid` → entitlement 有效  

自动化（9.2 / 9.3，含 `simulate-paid` 录制夹具）：

```powershell
pwsh docs/runbooks/merchant-console-e2e.ps1 `
  -BaseUrl "https://<gateway>" `
  -AdminUser admin `
  -AdminPassword "<hub-admin-password>"
```

前置：本变更已部署到目标网关；`CASH_PAYMENT_DEV_BYPASS=1`（simulate-paid）；Admin 账号可授期。SSE 推送仍建议浏览器打开 `/merchant` 目视一次。

## Flutter

兄弟仓 `flutter_ai_talk`：多宝宝列表/切焦点/创建虚拟宝宝（与本仓同发）。旧 App 仍「无→一」。兼容说明见兄弟仓 `docs/multi-baby-compat.md`。

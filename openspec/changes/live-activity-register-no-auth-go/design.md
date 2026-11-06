## Context

现状链路：

```
客户端 POST /app/api/push/live-activity/register + Bearer
  → gateway HookBeforeServe：InjectAccessHeadersFromBearer
       → 写 X-Internal-Wx-Id
  → 反代 push-service
  → LiveActivityRegisterPost：wxIDFromHeader
  → RegisterLiveActivityToken(wxID, …) → push_live_activity_token
```

Flutter twin 改为唤醒窗原生 POST、**无 Bearer**、body 带 `wxId`。gateway 未豁免则 401；豁免后若仍只读内部头，无头则失败。须双改：豁免 + body.`wxId`。

`StripSpoofedInternalHeaders` 已在鉴权分流前执行，豁免 path 上客户端伪造的 `X-Internal-Wx-Id` 会被剥掉——落库必须信 body，不能信客户端自带内部头。

## Goals / Non-Goals

**Goals:**

- 本 path 无 Bearer 可成功 upsert（body 非零 `wxId` + 既有字段合法）。
- 落库身份 = body.`wxId`；表与 upsert 键不变。
- 与 Flutter `live-activity-native-register-no-auth` 同 path、同字段表。

**Non-Goals:**

- 免鉴权扩散到 device push register/unregister。
- 改 APNs 扇出、1/min、pending 清场、Timing 钩子。
- 强制 IP/wxId 限流（可后续）。
- 修改 Flutter 仓代码。

## Decisions

1. **gateway：精确 POST 豁免本 path。**  
   写入 `gatewayAppAuthExemptExactPOST`：`/app/api/push/live-activity/register`。  
   豁免分支已有「若带了用户 Bearer 仍可尝试注入内部头」；无 Bearer 直接放行。  
   备选：整前缀豁免 `/app/api/push/` → 会放开 register/unregister，否决。

2. **push controller：身份改为 body.`wxId`。**  
   `AppPushLiveActivityRegisterPostReq` 增加 `WxId int64 \`json:"wxId"\``（校验 >0）。  
   `LiveActivityRegisterPost`：**不得**再 `wxIDFromHeader` 作为唯一来源；`wxID := req.WxId`，无效则 400。  
   若同时存在注入头与 body：以 body 为准（文档化），避免双源。

3. **同 path 原地加 `wxId`（已定，不走 api/v2）。**  
   与 Flutter 同 URL 在 `api/v1` 请求体增加 `wxId`（鉴权亦 BREAKING）。  
   依据：activity token register **尚未对用户发版**，不存在依赖「Bearer + 无 wxId + 头身份」的历史客户端；相对 AGENTS「v1 结构冻结」，本 path 视为未发布契约的首次定型，**不**另开 v2。  
   强制 body.`wxId`；**不**做「body 空时回退 `X-Internal-Wx-Id`」兼容。

4. **威胁模型（产品已定）。**  
   公开写面可污染某 `wxId` 的 LA token → 无法自动关锁屏；接受。不做签名/deviceKey 交叉校验（本 change）。

5. **RegisterLiveActivityToken 服务函数不变。**  
   仍 `(ctx, wxID, deviceKey, …)`；仅 controller 喂入的 wxID 来源变化。

## Risks / Trade-offs

- [伪造 wxId 覆写 token] → 产品接受；后续可限流。  
- [豁免过宽] → 仅精确 path，避免误伤。  
- [Go 先于 Flutter twin 合入] → 仅影响尚未发版的联调客户端；无线上历史流量依赖旧契约。  
- [AGENTS v1 冻结例外] → 见决策 3（未发版 path 首次定型）。

## Migration Plan

1. 与 Flutter twin `live-activity-native-register-no-auth` 同契约字段表；可同发或 Go 先合（无线上旧客户端）。  
2. 合入后：无 Bearer + body.`wxId` 即可 upsert；缺 `wxId` 一律参数错误。  
3. 回滚：移出豁免表 + 恢复 `wxIDFromHeader`；原生无 Bearer 路径再次 401。

## Open Questions

- （无）免 Bearer、body.`wxId`、同 path 不走 v2、不限流、不回退 header 均已定。

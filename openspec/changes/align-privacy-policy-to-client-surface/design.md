## Context

隐私政策托管于 Go 仓 `resource/public/privacy-policy.html`，App WebView / 官网同源访问。Flutter 侧能力开关（`ucg_feature_flags.dart`）与退役能力（VIP、care/growth、clinic 等）已使现网表面远小于现行文案。

基线：`openspec/specs/v3.2.0` → `privacy-policy-company`。Explore 已确认：无注册地址；微信不写头像昵称；商业化精简写；设备型号披露用途；隐藏 UGC；权限与 AI 对齐客户端。

## Goals / Non-Goals

**Goals:**

- 文案与现网用户可触达能力一致，不夸大、不残留已关模块。
- 保留必要合规项：运营主体、Apple 登录边界、喂养数据、推送、横屏语音 AI、精简商业化、设备型号。
- 规格与 HTML 同步，避免「规格仍要求社区/VIP/注册地址」。

**Non-Goals:**

- 不改 Flutter 功能开关或重新打开 UGC。
- 不改 `user-agreement.html`（本变更范围外）。
- 不删除服务端 UCG/旧 API 代码；仅文档对齐「客户端隐藏」产品面。
- 不新增 App HTTP 接口（无 usage 统计确认项）。

## Decisions

1. **对齐依据：客户端可触达表面，而非服务端仍存活的 API**  
   - UCG 主壳关闭 → 隐私政策不主动披露社区收集/位置/Green。  
   - 后端保留不影响本期文案。

2. **微信披露**  
   - 写：经微信授权获取用于账号识别的标识（unionid），用于登录。  
   - 明确：不收集微信头像、昵称。  
   - 不强调持久化 OpenID（实现仅落库 unionid）。

3. **商业化精简写**  
   - 一句：开通部分付费或邀请功能时，处理订单、支付渠道回执与功能权益状态。  
   - 不出现「VIP」字样；不列举已退役 care/growth。

4. **设备型号**  
   - 披露：客户端上报的设备型号展示字符串，用于崩溃排查与运维统计；非广告跟踪。  
   - 与 `add-wx-device-model-usage-column` / Flutter `report-device-model-to-server` 对齐。

5. **权限改写锚点（Flutter）**  
   - 相机：宝宝共享码扫一扫；记录配图；宝宝头像。  
   - 相册：相册识码；记录配图；宝宝头像。  
   - 麦克风：喂养/预测横屏语音。  
   - 删除位置权限段。

6. **第三方列表**  
   - 保留：智谱/DeepSeek/DashScope（语音 AI）、推送通道、应用内支付、对象存储（用户上传媒体，勿写「社区发帖」）。  
   - 删除或移出：阿里云 Green（社区审核）、以社区为唯一用途的 OSS 表述。

7. **注册地址**  
   - HTML 与规格均不再要求「以市场主体公示为准」句；主体全称 + 邮箱足够。

8. **修订日期**  
   - 更新为实施当日（或 2026-11-18 及之后）；生效日期可保留历史值。

## Risks / Trade-offs

- [Risk] 监管抽查认为「服务端仍有 UCG」却未披露 → Mitigation：产品面以客户端入口为准；UCG 重开时必须再改隐私政策与规格。  
- [Risk] 用户协议仍含旧表述 → Mitigation：本变更明示不改协议；可后续立 `align-user-agreement`。  
- [Risk] 精简商业化过短 → Mitigation：至少覆盖「订单/权益/支付渠道」三要素。

## Migration Plan

1. 合并后部署/刷新 `gateway-app-server` 静态资源。  
2. App 无需发版即可看到新文案（WebView 拉最新页；若有强缓存则清 CDN/浏览器缓存）。  
3. 回滚：还原 HTML 与规格 delta。

## Open Questions

（无；商业化精简写与设备型号用途已由产品确认。）

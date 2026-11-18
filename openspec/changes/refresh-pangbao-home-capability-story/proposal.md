## Why

官网 `pangbao-home.html` 仍按旧产品面叙事（同月龄轻社交、胖宝诊疗、值得留意、粉紫玻璃风），与现网能力及已修订的隐私政策不一致；用户协议仍含 UGC/微信头像昵称/注册地址等过时条款且视觉未跟进。需要一次改版：按现网能力重写官网故事与高档视觉，法律页同款风格，并让官网可匿名拉取小智引导视频。

## What Changes

- **重写官网信息架构与文案**（`resource/public/pangbao-home.html`）：优先级为智能预测/提醒 → 宝宝信息记录（事件 logo 叠放 + 示意统计图）→ 语音 AI 问答闭环 → 小智音箱场景与引导视频；Hero/能力/SEO 对齐现网；**官网 MUST NOT** 出现付费、会员、邀请开通等营销文案。
- **视觉升级**：浅色克制高档风（告别粉紫玻璃英雄感）；顶栏不 sticky；下载改为弹层；页脚压缩但仍含公司主体、ICP、法律链与邮箱。
- **小智引导视频公开只读**：官网匿名可拉取已配置槽（`promo`/`tutorial`）的 CDN 列表；既有 App Bearer 路径语义可保留或等价公开化（见 design）；实现前须向负责人确认是否计入 usage 统计。
- **法律页风格对齐官网**：`privacy-policy.html` 与 `user-agreement.html` 共用官网视觉 token（顶栏不 sticky、`a.brand` 无下划线、可读长文）。
- **用户协议内容对齐现网/隐私政策**：删除注册地址与 UGC 整节；微信改为 unionid、不收集头像/昵称；服务内容改为记录/统计/预测/语音 AI/小智等；知识产权去掉社区授权；**保留极简付费/开通条款**（协议合规；官网仍不写）；更新修订日期。
- **规格 MODIFIED**：`gateway-app-official-site`、`privacy-policy-company`（视觉 + 与现网一致的残留约束）、`user-agreement-company`、`xiaozhi-mcp-guide-videos`（匿名可读）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `gateway-app-official-site`：官网能力叙事、视觉、下载弹层、页脚压缩；去掉轻社交/诊疗主叙事；禁止付费营销文案。
- `privacy-policy-company`：法律页视觉须与官网同系；主体/披露要求与现网对齐（无注册地址强制；可承接已落地的客户端表面对齐）。
- `user-agreement-company`：内容对齐现网（无 UGC/无注册地址/微信 unionid）；保留极简付费与 AI 免责；视觉与官网同系。
- `xiaozhi-mcp-guide-videos`：已配置引导视频列表 MUST 支持官网匿名只读访问（Bearer App 路径可并存）。

## Impact

- **静态页**：`pangbao-home.html`、`privacy-policy.html`、`user-agreement.html`（均经 `gateway-app-server` 对外）。
- **API / 网关**：`cash-service` 引导视频读接口；`gateway_app_auth_exempt`（及必要时 usage skip）调整；可能新增或放行公开 GET。
- **规格**：基线 `openspec/specs/v3.2.0` 上述四 capability。
- **关联**：工作区可能已有 `align-privacy-policy-to-client-surface` 隐私正文改动；本变更以官网改版 + 用户协议内容 + 三页视觉统一为主，避免重复推翻已对齐正文。
- **不必改**：Flutter 法律页 URL；官网聚合 `/device/app/api/site/home` 数据契约可复用事件 logo（宝宝信息记录区），除非 Hero 文案字段需同步清洗。
- **部署**：`gateway-app-server`（静态页 + 鉴权白名单）；若改 cash 路由则另需 `cash-service`。

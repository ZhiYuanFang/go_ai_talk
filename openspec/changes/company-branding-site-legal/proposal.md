## Why

运营主体已由个人开发者变更为公司（杭州彩逗科技有限公司），应用商店账号已迁至公司名下，但官网页脚与隐私政策/用户协议仍呈「个人团队」观感，且能力披露落后于现网（预测提醒、推送、诊疗、付费开通等）。上架与对外展示需要统一公司主体、备案号与合规文案，增强正规感并满足审核对运营者信息披露的期望。

## What Changes

- 官网（`pangbao-home.html`）页脚展示公司中英文全称、ICP 备案号、客服邮箱；去掉运维向「权威数据」旁白；补充简短「关于我们 / 运营主体」信息。
- iOS 下载区改为可直达已上架的 App Store 链接（不再仅「搜索胖宝」）。
- 产品能力区适度对齐现网（预测提醒、胖宝诊疗、值得留意等），表述克制、保留 AI 免责声明。
- 隐私政策：开篇明确运营主体与联系方式；按现网能力补齐推送 token、预测临近同步、付费/功能开通、相关 AI 场景与第三方清单；更新生效日期；统一儿童/监护人说明。
- 用户协议：将「胖宝团队」改为公司主体；补充知识产权归属、注册地址表述（以市场主体公示为准）、付费与 AI 免责要点；与隐私政策主体一致。
- **非本仓范围（文档提示即可）**：App Store 列表页「开发者/版权」展示名由个人改为公司，属商店后台操作；本变更只保证官网与合规页与公司主体一致。

## Capabilities

### New Capabilities

- `company-site-legal`：官网页脚与主体展示、ICP/客服、App Store 直达、能力区与关于信息的对外展示要求。
- `privacy-policy-company`：隐私政策运营主体、联系方式及按现网能力的收集/第三方披露修订。
- `user-agreement-company`：用户协议运营主体、知识产权与付费/AI 相关条款修订。

### Modified Capabilities

- （无强制 delta）既有「Gateway-app 根路径承载胖宝官网」「隐私政策 SHALL 披露…」等要求已散落在归档版本规格中；本变更以新 capability 固化公司主体与本次文案范围，路径仍为 `/`、`/privacy-policy.html`、`/user-agreement.html`。

## Impact

- **静态页**：`resource/public/pangbao-home.html`、`privacy-policy.html`、`user-agreement.html`。
- **路由**：不改 gateway-app 路径；部署后随静态资源生效。
- **客户端**：加载既有合规 URL 即可看到更新；商店元数据迁移属账号后台。
- **常量（已确认）**：
  - 公司：杭州彩逗科技有限公司 / Hangzhou Caidou Technology Co., Ltd.
  - ICP：浙ICP备2024101366号
  - 客服：tou_zy@foxmail.com
  - App Store：https://apps.apple.com/cn/app/%E8%83%96%E5%AE%9D/id6774418472
  - 注册地址文案：以市场主体公示为准
  - 商店开发者账号：已迁公司（本变更不改商店后台）

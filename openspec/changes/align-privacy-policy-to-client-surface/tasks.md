## 1. 规格与文案修订

- [x] 1.1 按 design 大纲全文修订 `resource/public/privacy-policy.html`：删注册地址；改微信；精简商业化；加设备型号；去 UGC/位置；改权限；清洗 AI/第三方/不收集/注销
- [x] 1.2 更新页面「修订日期」为实施日；核对邮箱仍为 `pangbao@cuplay.top`、主体全称保留
- [x] 1.3 通读一遍：全文不得残留 VIP、社区发帖、属地/浏览位置、值得留意、成长轨迹、诊疗（现行）、Green 社区审核等矛盾表述
- [x] 1.4 顶栏 CSS：去掉 `position:sticky`（滚动不置顶）；`a.brand` 取消下划线

## 2. 验收

- [x] 2.1 对照 Flutter `kUcgHomePagerEnabled=false`、扫码/相册识码、横屏语音、feature 付费等现网能力做清单勾选
- [x] 2.2 确认路径仍为 `/privacy-policy.html`；Flutter 无需改 URL
- [x] 2.3 标明部署：需刷新/重部署 `gateway-app-server` 静态资源

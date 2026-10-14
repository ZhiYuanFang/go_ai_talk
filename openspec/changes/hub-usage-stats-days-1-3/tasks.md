## 1. 静态页时间范围预设

- [x] 1.1 `client-usage-stats.html`：`daysSelect` 增加近 1 天 / 近 3 天；顺序 1→3→7→30；默认 `selected` 在 1
- [x] 1.2 `api-usage-stats.html`：同等选项与默认近 1 天

## 2. 验收

- [x] 2.1 两页首开默认 days=1；可选 3/7/30；切换后请求带对应 `days`（代码按此；部署后目视）
- [x] 2.2 部署清单：gateway-app（静态）；无需改 voice/device/Python

# Flutter 兄弟仓手顺（本迭代）

本仓已交付 App 契约；UI 在 `flutter_ai_talk` 实现。

## 接口

- `GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）
- 响应：`{ list: [{ slot, title, videoCdnUrl, coverCdnUrl? }] }`
- 仅含已配置槽；空 `list` → **完全不渲染**引导入口
- **计入** gateway-app usage（勿当维护型接口跳过）

## 页面行为

1. 小智音箱页进入时拉取清单。
2. `list.isEmpty` → 无「了解能力 / 绑定教学」入口。
3. 有项 → 按 `title` 展示；点击用 `videoCdnUrl` 播放（可用 `coverCdnUrl` 作封面）。

## 联调

- 先在 Hub「小智 MCP 开通」区 D 上传合规 mp4（H.264+AAC，≤25MB）。
- App 登录后确认 GET 有数据、usage 面板可见该 path 计数。

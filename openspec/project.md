# 项目上下文

本文件是仓库 **唯一** AI / OpenSpec 项目入口（Cursor 规则与 `openspec/config.yaml` 均指向此处）。**不再维护** 根目录 `AGENTS.md`。行为级 MUST 以 `openspec/specs/<domain>/spec.md` 为准；下文只提供名片、布局与短指针。

## 项目目标

`go_ai_talk`：智能语音与设备管理后端，以 gateway + 领域服务（voice / device / history / ucg / cash / push / mcpbridge 等）演进微服务平台。客户端联调见兄弟仓 `flutter_ai_talk`。

## 技术栈

- Go / GoFrame（HTTP、路由、中间件、配置）
- Redis（缓存、临时态）、RabbitMQ（跨服务事件）
- Docker / Compose / GitHub Actions → ACR
- OpenSpec CLI 2.1.x（`spec-driven`：proposal → specs → design → tasks）

## OpenSpec 布局

- **真相**：`openspec/specs/<domain>/spec.md`（十八域，见下）。
- **变更**：`openspec/changes/<change-id>/`；delta 位于 `changes/<id>/specs/<domain>/spec.md`。
- **归档**：合并 delta 进主规格（若实现阶段已写妥主规格可 `--skip-specs`），然后 **删除**该 change 目录；默认 **不保留** `openspec/changes/archive/`。
- **制品语言**：简体中文；每条 `### Requirement` 须含英文 **SHALL** 或 **MUST**。
- **禁止**再使用 `openspec/specs/v*/` 巨石基线或 `hack/*openspec*` 版本合并脚本。

### 十八域

| 域 | 摘要 |
|---|---|
| `platform-runtime` | Redis/Rabbit/cachekit、配置边界、包隔离、一服务一库、导入门禁 |
| `platform-deploy` | Compose/Docker/CI/ACR、双栈环境、资源限制、cutover |
| `gateway-edge` | gateway-app/service 边缘、CORS/JWT、反代、官网页、usage 统计 |
| `admin-hub` | Hub/设备运维后台布局与管理 API |
| `observability` | 日志、背景循环治理、中文文档合规 |
| `identity-device` | 登录、微信/Apple、设备绑定、邀请、wx 画像 |
| `voice-chat` | ASR/LLM/TTS、流式对话、理解链路 |
| `history-feeding` | 喂养历史、事件层级、effective-day |
| `prediction` | 预测、临近推送、Live Activity、predict-pro |
| `care-commerce` | 护理留意点、现金 VIP、功能解锁与履约 |
| `clinic-bridge` | 胖宝诊疗、tip 反馈退役与宿主 |
| `ucg-social` | UCG 社交、OSS 媒体、审核/Feed |
| `push-notify` | 推送/APNs/HMS/notify 运行时 |
| `xiaozhi-mcp` | 小智 MCP 绑定、桥接、试用解锁 |
| `llm-lanes` | LLM 车道、配额、aimodel 退役 |
| `legal-site` | 隐私政策、用户协议、公司法律页 |
| `media-release` | APK/OSS CDN、App 状态横幅 |
| `merchant-console` | 商户 B 端台、虚拟宝宝配额/分发、商户 JWT/SSE、Vue 玻璃拟态面板（变更 `merchant-console-multi-baby` 归档后升主规格） |

## 工程约束指针

实现代码前对照对应域规格（勿在本文件重复长文 MUST）：

- 包边界 / 跨库禁止 / cachekit / 单行空结果 → `platform-runtime`
- CI/ACR / Compose / `*_DB_LINK` 发版清单 → `platform-deploy`
- gateway-app 反代 / auth exempt / usage 统计 → `gateway-edge`
- Hub 近全宽与左右分栏 → `admin-hub`
- 背景循环默认禁止 → `observability`
- `/opsx:apply` 收尾 MUST 列出需重部署的进程（或显式「不必重部署」）

## 重要约束（评审检查项）

- 行为变更是否落到某域 `### Requirement` / `#### Scenario`，禁止规格悬空。
- 是否误用已删除的 `v*.md` 巨石或 `hack/archive-openspec-to-version.py` 等脚本。
- 归档是否合并（或 skip-specs）并删除 change，而非堆 `changes/archive/`。
- 跨服务是否走契约（禁止跨库直查）；Redis 是否经 `cachekit`/`redismsgkit`。
- 新增微服务是否同步 CI/overlay/runbook/ACR；新 App 路由是否完成 gateway 登记与 usage 确认。

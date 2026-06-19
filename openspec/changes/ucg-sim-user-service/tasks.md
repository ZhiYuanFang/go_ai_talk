## 1. device 域：模拟用户标识与 internal API

- [x] 1.1 添加 `wx.is_simulated` 迁移与 entity/dao 字段（默认 0）
- [x] 1.2 实现 `POST /device/internal/api/sim/username/register`（internal 密钥 + `WxUsernameRegister` + `is_simulated=1`）
- [x] 1.3 实现 `GET /device/internal/api/sim/wx/list` 分页列表
- [x] 1.4 扩展 `POST /device/internal/api/ucg/wx/batch` 响应增加 `isSimulated`
- [x] 1.5 注册路由与 internal 中间件；补充 api/v1 契约与中文 summary

## 2. aimodel：sim Lane 与生图/生视频

- [x] 2.1 新增 Lane 常量 `simText`、`simVision`、`simImageGen`、`simVideoGen` 与默认 seed Profile
- [x] 2.2 实现 `GenerateImage`（CogView-3-Flash，Acquire/释放）
- [x] 2.3 实现 `SubmitVideoGeneration` 与 `PollVideoGeneration`（CogVideoX，仅 submit Acquire）
- [x] 2.4 智谱 video API HTTP 客户端（`/paas/v4/videos/generations`）与错误映射

## 3. ucg 域：internal 发消息

- [x] 3.1 定义 `POST /ucg/internal/api/chat/send` api/v1 契约
- [x] 3.2 实现 handler：校验 internal 密钥、device batch 确认 `is_simulated`、调用 `ProcessOutboundChatMessage`
- [x] 3.3 注册 ucg-service 路由（internal 中间件，与 media upload 一致）

## 4. gateway：usage 剔除与静态页

- [x] 4.1 `usagestats` 增加 `IsSimulatedWx`（Redis SET + device batch 回源）并在 `RecordAsync` 跳过
- [x] 4.2 注册 `/device/admin/sim-admin.html` 静态路由
- [x] 4.3 （可选）gateway 反代 sim Admin API 至 sim-user-service

## 5. sim-user-service 基础设施

- [x] 5.1 创建 `cmd/sim-user-service/main.go` 与 `manifest/config/config.sim-user-service.yaml`
- [x] 5.2 创建 sim DB 表：`sim_config`、`sim_prompt`、`sim_account_seq`、`sim_video_job`
- [x] 5.3 实现 HTTP 客户端：gateway（login + Bearer UCG API）、device internal、ucg internal
- [x] 5.4 启动时注册 aimodel ProfileStore；接入 Redis 与 `GLM_API_KEY`
- [x] 5.5 Docker compose 基线与 prod/test/local overlay 环境变量（`SIM_USER_SERVICE_ENABLED`、各任务开关、服务 URL）

## 6. sim-user-service 任务调度

- [x] 6.1 实现 scheduler 框架（独立 goroutine、ticker、±10% jitter、总开关与各任务开关）
- [x] 6.2 T1 register：序号分配、internal 注册、AI 昵称/头像、UCG profile/media
- [x] 6.3 T2 comment：随机 sim、feed、simVision 评论、POST comments
- [x] 6.4 T3 post_image：simText + simImageGen、media 链路、POST posts submit
- [x] 6.5 T4 post_video_submit：simText + SubmitVideoGeneration、写 sim_video_job（同 sim 单 pending 约束）
- [x] 6.6 P1 video_poll：1min 轮询、成功发帖/失败 skipped
- [x] 6.7 T5 chat_scan + E1 ephemeral window（30min×1min、真人 peer、去重 map）
- [x] 6.8 T6 follow：随机两 sim、POST follow 幂等
- [x] 6.9 media 上传对齐 UCG presign/register（extension、contentHash、transformVersion=sim-raw）
- [x] 6.10 extractChatContent 回退 reasoning_content（修复 post_image 等 simText 文案为空）

## 7. sim-user-service Prompt 运行时

- [x] 7.1 从 DB 加载 prompt 模板（6 种 taskType）与变量渲染（`{{post_content}}` 等）
- [x] 7.2 队列满/超时提前结束当前 tick 的统一错误处理

## 8. sim Admin API 与 Web

- [x] 8.1 实现 `GET/PUT /sim/admin/api/config`（enabled、maxSimUsers 默认 100）
- [x] 8.2 实现 `GET/PUT /sim/admin/api/prompts/{taskType}` 与种子默认 prompt
- [x] 8.3 实现 `GET /sim/admin/api/status`（任务上次运行、计数、pending jobs）
- [x] 8.4 编写 `resource/public/sim-admin.html`（配置、prompt 编辑、状态只读）

## 9. 文档与验收

- [x] 9.1 更新 `docs/runbooks/release-deploy-and-run.md`：sim-user-service 部署、开关、依赖 env
- [x] 9.2 更新 `manifest/docker/.env.example` 与 test/prod example 中 sim 相关变量
- [x] 9.3 运行 `openspec validate ucg-sim-user-service --strict` 并通过

### 9.4 前补丁：sim-admin 运维入口

- [x] 9.3.1 `resource/public/admin-modules.js`：增加 `sim-admin` 模块（`showInNav: true`，Hub 顶栏导航至 `/device/admin/sim-admin.html`，对齐 ucg-admin / voice-admin）
- [x] 9.3.2 `internal/controller/gateway_app_auth_exempt.go`：将 `GET/HEAD /device/admin/sim-admin.html` 加入 App Bearer 白名单（HTML 可加载，鉴权由页面内 `AdminCommon.requireAdmin()` 完成）
- [x] 9.3.3 `internal/services/gatewayapp/usagestats/skip.go`：`isStaticOrShellPath` 增加 `/device/admin/sim-admin.html`（与其它 Admin 壳页一致，不计入 usage）

- [ ] 9.4 测试环境手工验收：Hub 导航进入 sim-admin；注册 1 个 sim、发评论/图文、internal 发消息、usage 统计不含 sim wxId

## 10. CI 与部署闭环

- [x] 10.1 `.github/workflows/docker-acr.yml`：`ALL_SERVICES` 增加 `sim-user-service`；别名 `sim`/`sim-user`/`sim-user-service`；Dockerfile 映射；注释全量 7 服务
- [x] 10.2 `docker-compose.microservices.test.yml` / `prod.yml` / `local.yml`：sim-user-service 镜像、端口（test `19805`）、container_name
- [x] 10.3 `docker-compose.resources.test.yml` / `prod.yml`：sim-user-service mem/cpu limits
- [x] 10.4 `docs/runbooks/release-deploy-and-run.md`：六→七服务、清理脚本 grep、ACR 仓库清单、partial tag `+sim` 示例

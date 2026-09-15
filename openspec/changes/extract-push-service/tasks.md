## 1. 进程与库骨架

- [x] 1.1 新增 `cmd/push-service`、`config.push-service.yaml`、`internal/services/push` / `controller/push` 注册入口（addr 默认 `:9808`）
- [x] 1.2 建库/连库 `ai_voice_push`；实现 `push_device` 表 schema 确保（含唯一索引）；DAO/实体或等价访问
- [x] 1.3 从 ucg 迁入 APNs/HMS/MiPush 发送与 `PushByBizType`/`dispatch` 逻辑；配置改为 `PUSH_*` / `push.*`

## 2. API 与 gateway / usage

- [x] 2.1 App API：`POST /app/api/push/register`、`unregister`（登录 wxId）；Internal：`by-biz-type`（密钥 + 可选 badge，默认 0）
- [x] 2.2 gateway-app 反代 `/app/api/push/*` 至 push-service；鉴权与内部头与现 UCG App 路径对齐
- [x] 2.3 usage：`maintenance_skip` 加入新 register/unregister；删除旧 `/ucg/app/api/push/*` 排除项（负责人确认：不统计）

## 3. clients 与调用方迁移

- [x] 3.1 新增 `internal/clients/push`；更新 `clients/README` 若需要
- [x] 3.2 ucg：私信/评论/角标推送改 `clients/push`（带 badge）；删除 `push_*.go`、旧 App/Internal 推送路由与 `UCG_*` 推送 env 依赖
- [x] 3.3 voice 预测临近：改调 `clients/push`；更新 `predict-imminent` CLIENT/design 表述；删除 `clients/ucg` 推送封装（若仅为此存在）

## 4. 部署与 CI

- [x] 4.1 新增 `manifest/docker/Dockerfile.push-service`
- [x] 4.2 更新 `docker-compose.microservices.yml`（及 prod/local overlay、`.env.example`）：服务、URL、证书挂载、`PUSH_*`
- [x] 4.3 更新 `.github/workflows/docker-acr.yml`：ALL_SERVICES、别名、Dockerfile 映射
- [x] 4.4 更新 runbook / 部署文档：库名、端口、环境变量、无旧 path 兼容说明

## 5. 校验

- [x] 5.1 `go build` 相关 cmd；`hack/check-service-import` 通过（无业务包互引）
- [x] 5.2 自检：ucg 无厂商推送宿主残留；gateway 可到 push；usage skip 仅新 path

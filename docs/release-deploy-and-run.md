## 本版本部署与运行指南

适用于 change：`refactor-monolith-to-microservices-platform`（微服务拆分 + 分布式组件 + K8s 运维基线）。

### 1. 版本内容概览

本版本已完成：

- gateway / history-service / worker 三服务拆分与容器化
- Redis 集群、RabbitMQ、outbox 最终一致链路
- Kubernetes 基础清单（Deployment/Service/Ingress）
- HPA + PDB + RollingUpdate 发布策略
- 可观测基线（Prometheus/Loki/Tempo/Grafana）
- SLO、告警规则与故障响应手册

---

### 2. 安装 Docker 环境（Windows）

#### 2.1 自动安装（推荐）

在 PowerShell 中执行：

- `winget install -e --id Docker.DockerDesktop --accept-package-agreements --accept-source-agreements`

安装完成后：

1. 启动 Docker Desktop（首次启动会初始化 WSL2）
2. 如有提示，重启系统
3. 在新终端验证：

- `docker --version`
- `docker compose version`

#### 2.2 常见问题

- 提示 `docker` 命令不存在：
  - 先确认 Docker Desktop 已启动
  - 关闭并重新打开 PowerShell
- 首次启动报 WSL 相关错误：
  - 执行 `wsl --update`
  - 重启 Docker Desktop 后重试

---

### 3. 本地快速启动（Docker Compose）

#### 3.1 启动依赖组件

1. Redis 集群：

- `powershell -ExecutionPolicy Bypass -File "hack/redis-cluster-init.ps1"`

1. RabbitMQ（含交换机和队列初始化）：

- `powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`

#### 3.2 启动业务微服务

- `docker compose -f manifest/docker/docker-compose.microservices.yml up -d --build`

健康检查：

- gateway: [http://127.0.0.1:9701/api.json](http://127.0.0.1:9701/api.json)
- history-service: [http://127.0.0.1:9801/api.json](http://127.0.0.1:9801/api.json)
- worker: [http://127.0.0.1:9901/healthz](http://127.0.0.1:9901/healthz)

#### 3.3 启动可观测栈（可选）

- `docker compose -f manifest/docker/docker-compose.observability.yml up -d`

访问入口：

- Grafana: [http://127.0.0.1:3000](http://127.0.0.1:3000)（admin/admin）
- Prometheus: [http://127.0.0.1:9090](http://127.0.0.1:9090)

---

### 4. Kubernetes 部署（推荐学习路径）

#### 4.1 渲染与部署

1. 渲染 develop 清单：

- `kubectl kustomize manifest/deploy/kustomize/overlays/develop`

1. 应用到集群：

- `kubectl apply -k manifest/deploy/kustomize/overlays/develop`

1. 查看状态：

- `kubectl get deploy,svc,ingress,hpa,pdb -n default`

#### 4.2 环境化调整（values 等价）

在 `manifest/deploy/kustomize/overlays/develop` 调整：

- 镜像 tag：`kustomization.yaml` 的 `images`
- 副本与变量：`*-deployment.patch.yaml`
- 域名入口：`ingress.patch.yaml`
- 告警规则：`prometheus-rules.yaml`

---

### 5. 一致性与恢复验证

1. 触发 history 写入（新增/修改/删除记录）后，检查 outbox：

- `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ShowPendingOnly`

1. MQ 故障恢复后重放失败事件：

- `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ResetFailedToPending`

详细流程见：

- `docs/outbox-consistency-recovery.md`

---

### 6. 运维与回滚建议

- 告警与 SLO：`docs/slo-alerts-incident-runbook.md`
- 灰度与回滚（history 路由）：`docs/history-rollout-runbook.md`
- 可观测看板基线：`docs/observability-dashboards.md`

推荐上线顺序：

1. 先部署 history-service + gateway proxy（canary）
2. 再启用 worker 消费与 outbox relay
3. 最后启用 HPA/PDB 与告警规则


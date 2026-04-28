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

1. Redis 集群（Windows）：
- `powershell -ExecutionPolicy Bypass -File "hack/redis-cluster-init.ps1"`

2. Redis 集群（Linux）：
- `cd /www/wwwroot/go/go_ai_talk`
- `docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d`

3. RabbitMQ（Windows，含交换机和队列初始化）：
- `powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`

4. RabbitMQ（Linux）：
- `cd /www/wwwroot/go/go_ai_talk`
- `docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d`


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

### 4. Kubernetes 部署（一步一步）

> 本项目当前 K8s 清单主要部署业务服务（gateway/history/worker）。  
> Redis、RabbitMQ 可先继续用 Docker Compose 在宿主机运行，作为外部依赖接入。

#### 4.1 前置条件

1. 安装并验证 `kubectl`：
- `kubectl version --client`

2. 准备一个本地 K8s 集群（任选其一）：
- Docker Desktop Kubernetes
- minikube
- kind

3. 确认集群可用：
- `kubectl get nodes`

#### 4.2 准备镜像

方式 A（推荐学习期）：先在本机构建镜像，再推到你的镜像仓库。

1. 构建镜像：
- `docker build -f manifest/docker/Dockerfile.gateway-service -t <registry>/go-ai-talk/gateway:develop .`
- `docker build -f manifest/docker/Dockerfile.history-service -t <registry>/go-ai-talk/history-service:develop .`
- `docker build -f manifest/docker/Dockerfile.worker-service -t <registry>/go-ai-talk/worker:develop .`

2. 推送镜像：
- `docker push <registry>/go-ai-talk/gateway:develop`
- `docker push <registry>/go-ai-talk/history-service:develop`
- `docker push <registry>/go-ai-talk/worker:develop`

> 若用 minikube/kind，也可用其本地镜像加载方式替代 push。

#### 4.3 配置 overlay 镜像地址

编辑 `manifest/deploy/kustomize/overlays/develop/kustomization.yaml` 中 `images`：
- `name: go-ai-talk/gateway`
- `name: go-ai-talk/history-service`
- `name: go-ai-talk/worker`

把 `newName/newTag`（或 `newTag`）改成你的实际仓库与版本。

#### 4.4 部署到集群

1. 预览渲染结果：
- `kubectl kustomize manifest/deploy/kustomize/overlays/develop`

2. 应用清单：
- `kubectl apply -k manifest/deploy/kustomize/overlays/develop`

3. 查看资源状态：
- `kubectl get deploy,svc,ingress,hpa,pdb -n default`
- `kubectl get pods -n default -o wide`

4. 查看启动日志（排障）：
- `kubectl logs -n default deploy/gateway --tail=100`
- `kubectl logs -n default deploy/history-service --tail=100`
- `kubectl logs -n default deploy/worker --tail=100`

#### 4.5 访问服务

1. 先用端口转发验证：
- `kubectl port-forward -n default svc/gateway 9701:9701`

2. 本机访问：
- [http://127.0.0.1:9701/api.json](http://127.0.0.1:9701/api.json)

3. 若已安装 Ingress Controller，再通过 `ingress.patch.yaml` 配置的 host 访问。

#### 4.6 更新发布与回滚

1. 发布新版本：
- 更新 `kustomization.yaml` 的镜像 tag
- `kubectl apply -k manifest/deploy/kustomize/overlays/develop`

2. 观察滚动更新：
- `kubectl rollout status deploy/gateway -n default`
- `kubectl rollout status deploy/history-service -n default`
- `kubectl rollout status deploy/worker -n default`

3. 快速回滚（示例）：
- `kubectl rollout undo deploy/gateway -n default`

#### 4.7 环境化调整（values 等价）

在 `manifest/deploy/kustomize/overlays/develop` 调整：
- 镜像/tag：`kustomization.yaml` 的 `images`
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


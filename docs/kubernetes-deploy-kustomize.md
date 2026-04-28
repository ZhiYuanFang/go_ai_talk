## Kubernetes 部署清单（Task 7.2）

本阶段使用 `kustomize` 提供 `deployment/service/ingress` 与环境化镜像标签配置（作为 values 能力）。

### 目录

- base: `manifest/deploy/kustomize/base`
  - `gateway/history-service/worker` 三套 deployment + service
  - `ingress.yaml`
- overlay: `manifest/deploy/kustomize/overlays/develop`
  - 镜像 tag 覆盖（`develop`）
  - 副本数与环境变量 patch
  - ingress host 覆盖（`dev.go-ai-talk.local`）

### 使用方式

1. 渲染 develop 环境 YAML：
   - `kubectl kustomize manifest/deploy/kustomize/overlays/develop`
2. 部署到集群：
   - `kubectl apply -k manifest/deploy/kustomize/overlays/develop`
3. 查看资源：
   - `kubectl get deploy,svc,ingress -n default`

### 可调参数（values 等价）

- 镜像仓库/tag：
  - `overlays/develop/kustomization.yaml` 下 `images`
- 副本数：
  - `gateway-deployment.patch.yaml`
  - `history-deployment.patch.yaml`
  - `worker-deployment.patch.yaml`
- 域名入口：
  - `ingress.patch.yaml`

### 自动扩缩容与滚动发布（Task 7.3）

- HPA（`autoscaling/v2`）：
  - `gateway-hpa.yaml`：`min=1, max=5, cpu=70%`
  - `history-hpa.yaml`：`min=1, max=4, cpu=70%`
  - `worker-hpa.yaml`：`min=1, max=6, cpu=75%`
- PDB（`policy/v1`）：
  - `gateway/history-service/worker` 均设置 `minAvailable: 1`
- 滚动发布策略（Deployment）：
  - `maxSurge: 25%`
  - `maxUnavailable: 0`

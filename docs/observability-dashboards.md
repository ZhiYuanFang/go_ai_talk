## 可观测性基线（Task 7.4）

本阶段目标：为日志、指标、分布式追踪提供统一入口与最小可用看板能力。

### Kubernetes 清单侧

- 在 `gateway/history-service/worker` Deployment 中统一增加：
  - Prometheus 抓取注解（`prometheus.io/*`）
  - OTel 环境变量（`OTEL_SERVICE_NAME`、`OTEL_EXPORTER_OTLP_ENDPOINT`、`OTEL_EXPORTER_OTLP_INSECURE`）

### 本地可观测栈（Docker Compose）

- 文件：`manifest/docker/docker-compose.observability.yml`
- 组件：
  - Prometheus（`9090`）
  - Loki（`3100`）
  - Tempo OTLP gRPC（`4317`）+ 查询（`3200`）
  - Grafana（`3000`，admin/admin）
- 配置文件：
  - `manifest/docker/observability/prometheus.yml`
  - `manifest/docker/observability/tempo.yaml`

### 启动方式

1. 启动业务容器（可选）：
   - `docker compose -f manifest/docker/docker-compose.microservices.yml up -d`
2. 启动可观测栈：
   - `docker compose -f manifest/docker/docker-compose.observability.yml up -d`
3. 打开看板：
   - Grafana: [http://127.0.0.1:3000](http://127.0.0.1:3000)
   - Prometheus: [http://127.0.0.1:9090](http://127.0.0.1:9090)

### 验收清单

- [ ] 三个服务可被 Prometheus 抓取
- [ ] Grafana 可接入 Prometheus 指标源
- [ ] 服务可向 Tempo 上报追踪数据
- [ ] Loki 可接入日志并支持按服务过滤

## 1. 服务专属配置落地

- [x] 1.1 新增 `config.voice-service.yaml` 与 `config.device-service.yaml`，迁移各自专属配置项。
- [x] 1.2 更新 `cmd/voice-service/main.go`、`cmd/device-service/main.go` 默认配置加载路径，对齐 `history-service` 模式。
- [x] 1.3 校验各服务环境变量覆盖规则（DB link、监听地址等）仅作用于本服务。

## 2. 主配置瘦身与边界清理

- [x] 2.1 盘点 `manifest/config/config.yaml` 中的服务专属字段并迁移到对应专属配置文件。
- [x] 2.2 删除主配置中不该由 gateway/全局承载的领域配置项，确保主配置职责单一。
- [x] 2.3 校验 gateway 启动不再依赖已迁移的服务专属配置字段。

## 3. 运行配置与部署清单对齐

- [x] 3.1 更新 docker-compose 微服务清单中各服务 `GF_GCFG_FILE` 与覆盖环境变量。
- [x] 3.2 更新 kustomize base/overlay 中各服务配置路径与覆盖变量，确保多环境一致。
- [x] 3.3 补充配置迁移说明与回滚策略，覆盖 local/canary/remote 常见切换场景。

## 4. 边界剥离收口

- [x] 4.1 检查并修复仍通过主配置耦合他域职责的代码路径（含跨服务访问入口）。
- [x] 4.2 对齐“跨服务走接口，不跨库直查”的配置与实现约束，避免回流路径。
- [x] 4.3 更新仓库级治理文档与评审检查项，固化最终态微服务边界规范。

## 5. 验证与验收

- [ ] 5.1 验证 voice/device/history/gateway 在专属配置下均可独立启动并通过健康检查。
- [ ] 5.2 验证主配置瘦身后系统关键链路功能无回归。
- [ ] 5.3 验证切换失败时可按服务维度回滚到旧配置并恢复可用。

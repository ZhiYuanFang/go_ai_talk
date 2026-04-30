## 1. 领域路由中间件拆分

- [x] 1.1 将现有 voice/device 代理逻辑从同一文件拆分为两个独立中间件实现。
- [x] 1.2 为 voice 与 device 分别定义独立配置结构与初始化流程（包含 once 缓存与配置读取）。
- [x] 1.3 更新 gateway 路由注册流程，按领域安装独立中间件并保持行为兼容。

## 2. voice/device canary 能力对齐

- [x] 2.1 为 voice 路由引入 `local|proxy|canary` 三态模式与 canary 百分比配置。
- [x] 2.2 为 device 路由引入 `local|proxy|canary` 三态模式与 canary 百分比配置。
- [x] 2.3 为 voice/device 增加稳定分流键计算与无状态百分比分流逻辑。

## 3. 配置与部署同步

- [x] 3.1 更新 `manifest/docker/docker-compose.microservices.yml`，补齐 voice/device canary 变量。
- [x] 3.2 更新 kustomize gateway 部署清单，补齐 voice/device canary 变量。
- [x] 3.3 校验非法模式、非法百分比、空目标地址时的回退语义符合预期。

## 4. 验证与文档收敛

- [ ] 4.1 执行 voice/device 路由冒烟验证（local、proxy、canary 三种模式）。
- [ ] 4.2 验证同一分流键在 canary 下命中稳定（不在 local/proxy 之间抖动）。
- [x] 4.3 更新网关端点归属矩阵与契约文档，明确 voice/device 已对齐 history 的治理模式。

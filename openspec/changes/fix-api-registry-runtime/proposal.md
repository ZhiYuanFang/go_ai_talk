## Why

App API 使用统计页已有调用次数，但「中文说明」列全部为「未登记」。`api/v1` 源码中已定义 `g.Meta` 的 `summary`（如「推荐 Feed」「点赞」），但 `apiregistry` 在 **gateway-app Docker 运行时** 通过 `gfile.MainPkgPath()` + 磁盘扫描 `api/v1` 加载；镜像未包含 `api/v1` 目录，registry 为空，读写两侧的 `Normalize` / `SummaryOf` 均无法命中。需在容器/生产环境可靠加载路由元数据，并补齐 summary 展示与模板匹配。

## What Changes

- `apiregistry` 改为 **`go:embed api/v1`** 编译期嵌入 Meta 源文件，运行时解析（不依赖镜像内 `api/v1` 目录）
- 保留开发环境磁盘扫描作 fallback（本地 `go run` 未 rebuild 时仍可读最新 Meta）
- **`SummaryOf`** 在 exact 未命中时调用 **`matchTemplate`**（与 `Normalize` 一致），兼容 Redis 中已存在的 raw id 路径 apiKey
- **`Dockerfile.gateway-app-service`** 可选 `COPY api` 双保险（非主路径，embed 为主）
- 不修改 Redis 键结构；部署后新写入将正确模板化 apiKey，历史 raw key 仍可通过 template 匹配显示 summary

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `gateway-app-api-usage-stats`：管理端 list/detail/user 返回的 `summary` MUST 在 `api/v1` 已登记时展示中文说明，不得因 registry 未加载而全部为「未登记」

## Impact

- `internal/services/gatewayapp/apiregistry/registry.go` — embed 加载 + SummaryOf 模板匹配
- `manifest/docker/Dockerfile.gateway-app-service` — 可选 COPY api（双保险）
- 需 **gateway-app** 镜像重建部署
- 不涉及 device/ucg/history/voice 服务变更

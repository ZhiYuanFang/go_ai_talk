## 1. apiregistry 编译期嵌入

- [x] 1.1 `apiregistry/registry.go`：`go:embed api/v1`（或等价递归 embed），`loadFromAPIV1` 从 embed FS 解析 `g.Meta`，移除对 `gfile.MainPkgPath()` 磁盘路径的依赖
- [x] 1.2 确认 embed 后 `Init()` 在 gateway-app 进程内 routes 非空（可 log 或 build 期 sanity）

## 2. SummaryOf 模板匹配

- [x] 2.1 `SummaryOf`：exact miss 时解析 apiKey 为 METHOD+path，调用 `matchTemplate` 返回 summary
- [x] 2.2 与 `Normalize` 行为对齐；空 summary 仍返回「未登记」

## 3. Docker 双保险（可选）

- [x] 3.1 `Dockerfile.gateway-app-service`：`COPY api ${WORKDIR}/api`（便于排查；embed 为主路径）

## 4. 校验

- [x] 4.1 `openspec validate fix-api-registry-runtime --strict`
- [x] 4.2 `go build ./...`
- [x] 4.3 部署 gateway-app 后：usage list 中 Feed/点赞等已登记 API 的 summary 非「未登记」

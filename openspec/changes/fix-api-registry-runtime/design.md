## Context

- `apiregistry` 从 `api/v1/**/*.go` 解析 `g.Meta` 的 `path`/`method`/`summary`，供 usage 统计 **apiKey 归一化**与运维页 **中文 summary**。
- 当前实现：`loadFromAPIV1()` 使用 `filepath.Join(gfile.MainPkgPath(), "api", "v1")`  Walk 磁盘；编译后 `MainPkgPath()` 常为空，fallback `"."`；Docker 镜像 WORKDIR=`/app` 且无 `api/v1` → **registry 为空**。
- `api/v1` 源码已含 App 路由 summary（如 `ucg_app_http.go`）；问题在 **运行时加载**，非缺登记文案。
- `SummaryOf` 仅 `byExact` 查找，不做 `matchTemplate`；registry 空时写入侧 `Normalize` 也失败，Redis 可能存 raw id 路径。

## Goals / Non-Goals

**Goals**

- gateway-app 容器内 `apiregistry.Init()` 后 registry 非空，已登记 App 路由能展示中文 summary
- `SummaryOf(apiKey)` 对模板 apiKey 与可匹配的 raw 路径均能返回 summary
- 写入侧 `Normalize` 在容器内正确模板化 apiKey（新数据）

**Non-Goals**

- 修改 `api/v1` 中已有 summary 文案（除非发现明显缺失，另开任务）
- Redis 历史 raw apiKey 合并/迁移（靠 template 匹配展示即可）
- 其它服务镜像的 apiregistry（仅 gateway-app 使用）

## Decisions

### 1. 主路径：`go:embed api/v1`

- 在 `apiregistry` 包内 `//go:embed api/v1/*.go`（或 `all:` 递归）嵌入 Meta 源文件
- `loadFromAPIV1` 优先从 embed FS Walk/Read，解析逻辑与现有一致（按行扫 `g.Meta`）
- **理由**：不依赖镜像 COPY、不依赖 `MainPkgPath`；与二进制同生命周期，生产可靠
- **备选**：仅 Dockerfile `COPY api` — 易漏、路径耦合 WORKDIR，弃为主方案

### 2. 开发 fallback：磁盘扫描保留

- embed 加载完成后，若 `gfile.MainPkgPath()` 非空且 `api/v1` 存在，可选合并磁盘扫描（后加载覆盖或 skip duplicate key）
- **或** simpler：embed 为主，磁盘仅当 embed 解析结果为 0 条时尝试（避免 dev 热改未 rebuild 时 stale — 实际上 dev 也会 rebuild）
- **采用**：**仅 embed** 作为唯一来源（`api/v1` 已在 repo，build 时 embed）；去掉对 MainPkgPath 的依赖，简化逻辑

### 3. `SummaryOf` 补 `matchTemplate`

- exact miss 时：`matchTemplate(method, pathFromApiKey)` → 返回 best.Template 的 summary
- apiKey 格式为 `METHOD /path`；与 `Normalize` 行为对齐
- **理由**：Redis 中历史 `POST /ucg/app/api/posts/123/like` 仍可显示「点赞」

### 4. Dockerfile 双保险（可选）

- `COPY api ${WORKDIR}/api` — 低优先级；embed 已足够。若加 COPY 不影响 embed，可作运维排查便利

## Risks / Trade-offs

- **[Risk] embed 路径与 module 布局变化** → embed 指令与 `api/v1` 同 repo 固定路径；CI `go build` 验证
- **[Risk] 新增 api 文件未触发 gateway-app rebuild** → 与现有部署流程一致，无额外风险
- **[Trade-off] 二进制略增** → `api/v1` 纯 Go 文本，体积可忽略

## Migration Plan

- 重建部署 **gateway-app** 即可；无需 Redis 迁移
- 部署后：运维页刷新，summary 应显示中文；新请求写入模板化 apiKey
- 回滚：回退镜像，summary 再次「未登记」，数据不受影响

## Open Questions

- 无；若统计页仍有个别「未登记」，再补 `api/v1` 缺失的 `g.Meta` 条目（独立小 PR）

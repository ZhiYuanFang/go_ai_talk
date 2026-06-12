## Context

- 现有 workflow：push `v*` tag → `resolve-env` 解析 test/prod 与 `primary_tag` → 固定 6 项 matrix 并行 build-push。
- 六服务 id：`gateway`、`gateway-app`、`history-service`、`voice-service`、`device-service`、`ucg-service`。
- 用户明确：**不要 retag**；缺镜像是正确信号；部署时 **只 pull 改动的服务**。

## Goals / Non-Goals

**Goals:**

- Tag `v1.0.0-rc.3+ucg` 仅构建 `ucg-service`，push `:v1.0.0-rc.3` 与 `:${GITHUB_SHA}`。
- 无 `+` 后缀 → 全量 6 服务（向后兼容）。
- `primary_tag`、test/prod 判定使用 **`+` 前 base tag**（`.env` 的 `IMAGE_TAG` 不含 `+ucg`）。
- `workflow_dispatch` 支持可选 `services` 输入。

**Non-Goals:**

- Retag / 镜像 manifest 复制。
- 按 git diff 自动推断服务（后续可另开变更）。
- 修改 compose overlay 为多服务独立 `IMAGE_TAG`。
- Docker BuildKit GHA cache（可后续叠加）。

## Decisions

### 1. Tag 解析

```bash
tag_raw="v1.0.0-rc.3+ucg,gateway"   # git tag 全名
base_tag="${tag_raw%%+*}"            # v1.0.0-rc.3 → primary_tag / IMAGE_TAG
scope="${tag_raw#*+}"                # 无 + 时 scope 等于 tag_raw，需判等处理

# deploy_env、semver 路由：resolve_env_from_tag("${base_tag}")
```

- 无 `+`：`build_all=true`，matrix = 6 服务。
- 有 `+`：解析 scope 为别名列表 → 映射 matrix id；非法别名 **fail workflow**。

### 2. 服务别名表（tag / dispatch 输入 → matrix id）

| 别名 | matrix id |
|------|-----------|
| `gateway` | gateway |
| `gateway-app` | gateway-app |
| `history` / `history-service` | history-service |
| `voice` / `voice-service` | voice-service |
| `device` / `device-service` | device-service |
| `ucg` / `ucg-service` | ucg-service |
| `all` | 全量 6 服务（显式） |

多服务：`+ucg,gateway`（逗号分隔，trim 空白）。

### 3. 不 retag

- 部分构建 **仅 push** 选中服务的 `:${primary_tag}` 与 `:${GITHUB_SHA}`。
- 未选服务在 ACR **可能不存在** `:${primary_tag}` — 符合预期。
- Workflow **成功** 不要求六仓库 tag 齐全。

### 4. 部署语义（runbook）

部分发版推荐流程：

```bash
# .env IMAGE_TAG=v1.0.0-rc.3（与 git tag + 前 base 一致）
docker compose -f ... pull ucg-service
docker compose -f ... up -d --no-build ucg-service
```

- **勿** 在无全量镜像时对全栈 `compose pull`（会因缺 gateway 等 tag 失败 — 用户视为配置错误）。
- 未重建服务继续运行 **已拉取的旧 tag 本地镜像**，直到下次全量 tag 发版。

### 5. workflow_dispatch

新增可选 input `services`（逗号分隔，空=全量），与 tag 后缀共用解析函数；`image_tag` 仍填 base tag（不含 `+`）。

### 6. resolve-env 输出

新增 outputs：

- `build_matrix_json`：`[{"id":"ucg-service","dockerfile":"..."}]`
- `build_scope`：`all` 或 `partial`

`build-push` 使用 `matrix: ${{ fromJson(needs.resolve-env.outputs.build_matrix_json) }}`。

## Risks / Trade-offs

- **[Risk] 误将 IMAGE_TAG 改为新 rc 后全栈 pull** → runbook 强调按服务 pull；失败即提示检查是否用了 `+` 部分 tag。
- **[Risk] 生产 tag `v1.0.0+ucg`** → 技术上允许，但仅 push 部分生产镜像；文档建议 **预发布 rc 场景使用 `+` 后缀**，生产全量发版用无后缀 tag。
- **[Trade-off] git tag 名与 IMAGE_TAG 不完全一致**（tag 含 `+ucg`，IMAGE_TAG 不含）→ runbook 必须写清，避免运维困惑。

## Migration Plan

- 合并 workflow + runbook 后即生效；现有无后缀 tag 行为不变。
- 无需 ACR / ECS 配置变更。

## Open Questions

- 无（retag 明确不做）。

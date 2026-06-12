## 1. Workflow 解析与动态 matrix

- [x] 1.1 在 `resolve-env` 解析 git tag：拆分 `base_tag`（`+` 前）与 `build_scope`（`+` 后）；`primary_tag` 输出 base tag
- [x] 1.2 实现服务别名 → matrix id 映射；非法别名 fail；无 `+` 或 `all` → 全量 6 服务
- [x] 1.3 输出 `build_matrix_json`；`build-push` 改为 `fromJson` 动态 matrix（保留各服务 dockerfile 映射）
- [x] 1.4 `workflow_dispatch` 增加可选 `services` 输入，与 tag 后缀共用解析逻辑

## 2. 文档

- [x] 2.1 更新 `.github/workflows/docker-acr.yml` 头部注释： `+ucg` 示例、无 retag 说明
- [x] 2.2 更新 `docs/runbooks/release-deploy-and-run.md`：部分构建 tag、IMAGE_TAG 与 git tag 区别、按服务 pull/up、全栈 pull 失败预期

## 3. 校验

- [x] 3.1 本地检查 workflow YAML 语法（或 push 前人工 review matrix JSON 输出 logic）
- [x] 3.2 `openspec validate docker-acr-tag-build-scope --strict` 通过

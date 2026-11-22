---
name: /opsx-archive
id: opsx-archive
category: Workflow
description: 归档已完成的 OpenSpec change——合并 delta 进域规格后删除该 change（不保留 archive 历史，不用 v* 巨石）
---

本仓库 **归档单个 change**（对齐 flutter_ai_talk；**不是**按版本号合并巨石基线）。

**禁止**：`hack/archive-openspec-to-version.py`、写入 `openspec/specs/vX.Y.Z/spec.md`、或把参数当成版本号收版。

**Input**：可选 change 名（如 `/opsx-archive add-auth`）。若省略或含糊，列出活跃 changes 请用户选择。

**Steps**

1. **选择 change**  
   `openspec list --json`。勿猜测；请用户确认。

2. **检查产物与 tasks**  
   `openspec status --change "<name>" --json`；读 `tasks.md` 未完成项并确认是否仍要归档。

3. **评估 delta 同步**  
   对比 `openspec/changes/<name>/specs/<domain>/spec.md` 与 `openspec/specs/<domain>/spec.md`。  
   选项：立即合并 / 跳过合并（主规格已更新）/ 取消。

4. **执行归档（合并后删除，不留 archive）**  
   - 合并 delta 进主规格，或确认已含变更后 `openspec archive <name> -y --skip-specs`（仅作辅助）。  
   - **删除** `openspec/changes/<name>/`。  
   - 若 CLI 写入 `openspec/changes/archive/YYYY-MM-DD-<name>/`，**一并删除**。  
   - 默认 **不保留** dated archive。

5. **校验**  
   `openspec validate --specs --strict`；确认该 change 已不在 `openspec list`。

6. **摘要**  
   报告 change 名、是否合并 specs、已删除路径。

## Success output

```
## Archive Complete

**Change:** <change-name>
**Specs:** merged | skipped
**Deleted:** openspec/changes/<change-name>/
```

## Guardrails

- 参数是 change-id，不是 `vX.Y.Z`
- 不创建/不保留 `openspec/specs/v*`
- 不调用已删除的 `hack/*openspec*` 脚本

---
name: openspec-archive-change
description: 按版本号收版——将活跃 OpenSpec changes 合并进目标版本基线并清空 changes。当用户说 /opsx-archive、收版、archive 跟版本号（如 v3.0.3）时使用。
license: MIT
compatibility: Requires openspec CLI and hack/archive-openspec-to-version.py.
metadata:
  author: openspec
  version: "2.0"
  generatedBy: "1.3.0"
---

# 按版本收版（本仓库）

**本仓约定**：`/opsx-archive vX.Y.Z` = 把 `openspec/changes/` 全部活跃变更收进 `openspec/specs/vX.Y.Z/spec.md`，**不**保留 `openspec/changes/archive/`。

**禁止**：把参数当成单个 change-id，去 `mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>`。

## Input

- **必须**提供目标版本：`vX.Y.Z`（例：`v3.0.3`）
- 可选第二参数：源基线版本（默认取比目标小的最新已有版本）
- 若未给版本号或格式非法：列出 `openspec/specs/v*/` 与活跃 changes，请用户确认目标版本

## Steps

1. **列出将收版的 changes**  
   `openspec list --json`。未完成 tasks 的 change 默认一并收版；仅当用户明确要求时才排除。

2. **执行权威脚本**

   ```bash
   python hack/archive-openspec-to-version.py vX.Y.Z
   ```

   脚本：`split` 源基线 → 逐个 `openspec archive --yes` → **删除** archive 日期目录 → `merge` 到 `openspec/specs/vX.Y.Z/spec.md` → 清理临时 capability 目录。

3. **补救残留**  
   若 `openspec list` 仍有 change（常见原因：MODIFIED/REMOVED 标题与基线不一致，CLI 可能 Aborted 仍 exit 0）：
   - 修正 delta 中 Requirement 标题与当前基线一致后重试 `openspec archive <name> --yes`
   - 删除对应 `archive/YYYY-MM-DD-*`
   - 再 `python hack/merge-openspec-specs.py vX.Y.Z` 并清理非版本 capability 目录
   - 仅在用户明确同意时用 `--skip-specs`

4. **更新 `openspec/project.md`**  
   基线引用改为 `openspec/specs/vX.Y.Z/spec.md`。

5. **校验**  
   `openspec validate --strict`；确认活跃 changes 已空（或仅剩用户保留项）。

6. **摘要**  
   报告目标版本、源基线、spec 路径、残留（若有）。

## Success output

```
## 收版完成

**目标版本:** vX.Y.Z
**源基线:** vA.B.C
**规格:** openspec/specs/vX.Y.Z/spec.md
**活跃 changes:** 已清空
**project.md:** 基线引用已更新
```

## Guardrails

- 参数是版本号，不是 change 名
- 不保留 archive 日期目录
- 权威入口：`hack/archive-openspec-to-version.py`
- 命令文档：`.cursor/commands/opsx-archive.md`
- 总览：`.cursor/skills/openspec/SKILL.md`「收版」

---
name: /opsx-archive
id: opsx-archive
category: Workflow
description: 按版本号收版——将活跃 OpenSpec changes 合并进目标版本基线并清空 changes
---

本仓库 **收版**（不是上游「归档单个 change」）。

将 `openspec/changes/` 下全部活跃变更合并进目标版本基线 `openspec/specs/vX.Y.Z/spec.md`，**不保留** `openspec/changes/archive/` 目录内容。

**Input**：必须提供目标版本号，例如 `/opsx-archive v3.0.3` 或 `/opsx:archive v3.0.3`。

- 合法格式：`v` + 三段数字（`vX.Y.Z`）
- **禁止**把参数当作单个 change-id（如 `add-auth`）去 `mv` 到 `archive/YYYY-MM-DD-*`
- 若未提供版本号或格式非法：列出已有 `openspec/specs/v*/` 与当前活跃 changes 数量，请用户给出目标版本

**Steps**

1. **确认目标版本**

   - 解析参数为 `vX.Y.Z`
   - 默认源基线：比目标小的最新已有版本（脚本自动推断；也可显式传第二参数）
   - 用 `openspec list --json` 展示将要收进本版的活跃 change 列表（含未完成 tasks 的也一并收版，除非用户明确要求先排除）

2. **执行收版脚本（权威入口）**

   ```bash
   python hack/archive-openspec-to-version.py vX.Y.Z
   # 可选显式源基线：
   # python hack/archive-openspec-to-version.py vX.Y.Z v3.0.0
   ```

   脚本流程（勿手写替代，除非脚本失败需补救）：
   - `split-openspec-baseline.py` 拆开源基线 → capability 目录
   - 对每个活跃 change 执行 `openspec archive <name> --yes`
   - **删除** `openspec/changes/archive/YYYY-MM-DD-<name>/`（本仓约定不保留 archive）
   - `merge-openspec-specs.py` 合并为 `openspec/specs/vX.Y.Z/spec.md`
   - 清理临时 capability 目录

3. **处理脚本中途失败（MODIFIED/REMOVED 标题对不上等）**

   - `openspec archive` 可能在 specs 应用 Aborted 时仍返回 0，导致 change 目录残留
   - 核对 `openspec list`：若仍有残留，先 `split` 当前目标版本（若已部分 merge）或源版本，修正 delta 中 Requirement 标题与基线一致后重试 `openspec archive <name> --yes`
   - 成功后删除对应 `archive/YYYY-MM-DD-*`，再 `merge-openspec-specs.py vX.Y.Z` 并清理 capability 目录
   - 仅在用户明确同意时才对单个 change 使用 `--skip-specs`

4. **更新基线引用**

   将 `openspec/project.md`（及本仓 OpenSpec skill 示例若仍写旧版本）中的基线路径从上一版改为 **`openspec/specs/vX.Y.Z/spec.md`**。

5. **校验**

   ```bash
   openspec validate --strict
   openspec list   # 期望无活跃 change（或仅剩用户明确保留的）
   ```

6. **展示摘要**

**Output On Success**

```
## 收版完成

**目标版本:** vX.Y.Z
**源基线:** vA.B.C
**规格:** openspec/specs/vX.Y.Z/spec.md
**活跃 changes:** 已清空（或不保留 archive）
**project.md:** 基线引用已更新为 vX.Y.Z

已跑 openspec validate --strict。
```

**Output On Partial Failure**

```
## 收版部分完成（有残留）

**目标版本:** vX.Y.Z
**已合并:** openspec/specs/vX.Y.Z/spec.md（若已 merge）
**残留 changes:** <name1>, <name2>
**原因:** openspec archive 应用 delta 失败（常见：MODIFIED/REMOVED 标题与基线不一致）

下一步：修正 delta 标题后重试 archive，或经确认后 --skip-specs。
```

**Guardrails**

- **本命令 = 按版本收版**；禁止实现为「只归档某一个 change 到 archive/」
- 不保留 `openspec/changes/archive/` 下的日期目录（脚本会删；补救路径也须删）
- 参数必须是 `vX.Y.Z`，不是 change-id
- 收版后必须更新 `openspec/project.md` 基线引用
- 未完成 tasks 的 change 默认一并收版；若用户要求保留未完成项，先移出/排除再跑脚本
- 权威实现：`hack/archive-openspec-to-version.py`；说明见 `.cursor/skills/openspec/SKILL.md`「收版」

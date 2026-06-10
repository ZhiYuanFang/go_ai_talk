---
name: openspec-archive-change
description: 将活跃 OpenSpec 变更收进目标版本基线（不保留 archive 目录）。Use when the user wants to finalize changes and merge specs into a version baseline like v2.0.3.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "2.0"
  generatedBy: "1.3.0"
---

将活跃 change 的 delta spec 合并进指定版本基线 `openspec/specs/<version>/spec.md`，**不**保留 `openspec/changes/archive/`。

**Input（必填）**：目标版本名，如 `v2.0.3`。可选源基线版本（默认低于目标的最新 `vX.Y.Z`）。

**Steps**

1. **解析版本**

   从用户消息提取 `vX.Y.Z`。未提供则提示用户给出版本名，不要猜测。

2. **确认源基线与活跃变更**

   - 源基线：默认 `openspec/specs/` 下低于目标的最新版本
   - `openspec list --json` 列出活跃 change
   - 对每个 change 检查 artifact（`openspec status --json`）与 `tasks.md` 未完成项；有警告时确认是否继续

3. **执行收版**

   ```bash
   python hack/archive-openspec-to-version.py <target> [source]
   ```

   手工等价流程：
   1. `python hack/split-openspec-baseline.py <source>`
   2. 对每个 change：`openspec archive <name> --yes` → 立即删除 `openspec/changes/archive/*-<name>/`
   3. `python hack/merge-openspec-specs.py <target>`
   4. 删除 `openspec/specs/` 下非 `vX.Y.Z` 的 capability 目录

   **依赖顺序**：先归档仅 ADDED 新 capability 的 change，再归档含 MODIFIED 的 change。REMOVED 头不匹配时修正 delta 后重试。

4. **更新 project.md**

   将 `openspec/project.md` 基线引用更新为目标版本。

5. **摘要**

   报告目标版本、源基线、已收版 change 列表、警告项、`openspec/specs/<target>/spec.md` 路径。

**Guardrails**

- 禁止保留 archive 目录与散落 capability specs
- 必须先 split 源基线再 archive（MODIFIED 需要已有 capability spec）
- 收版后更新 `openspec/project.md`

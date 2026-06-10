---
name: /opsx-archive
id: opsx-archive
category: Workflow
description: 将活跃 OpenSpec 变更收进目标版本基线（不保留 archive 目录）
---

将活跃 change 的 delta spec 合并进指定版本基线文件 `openspec/specs/<version>/spec.md`。

**不**创建 `openspec/changes/archive/` 目录；change 目录在 spec 同步后由 `openspec archive` 移除。

**Input（必填）**：目标版本名，如 `v2.0.3`。

示例：`/opsx-archive v2.0.3`

可选第二参数：源基线版本（默认取 `openspec/specs/` 下低于目标版本的最新 `vX.Y.Z`）。

**Steps**

1. **解析版本**

   从用户消息提取目标版本（须匹配 `vX.Y.Z`）。若未提供，**必须**提示用户给出版本名（如 `v2.0.3`），不要猜测。

2. **确认源基线**

   列出 `openspec/specs/v*/` 已有版本；默认源基线为低于目标的最新版本（如目标 `v2.0.3` → 源 `v2.0.2`）。若目标目录已存在，询问是否覆盖或取消。

3. **检查活跃变更**

   运行 `openspec list --json`。若无活跃 change，仍可仅基于当前 capability specs 执行 merge（少见）。

   对每个活跃 change：
   - 运行 `openspec status --change "<name>" --json` 检查 artifact
   - 读取 `tasks.md` 统计未完成项 `- [ ]`
   - 若有未完成 artifact/task，显示警告并确认是否继续

4. **执行收版（推荐脚本）**

   ```bash
   python hack/archive-openspec-to-version.py <target-version> [source-version]
   ```

   脚本等价步骤：
   - `python hack/split-openspec-baseline.py <source>` — 从源基线拆出 capability specs
   - 对每个活跃 change：`openspec archive <name> --yes`，随后**立即删除** `openspec/changes/archive/*-<name>/`
   - `python hack/merge-openspec-specs.py <target>` — 合并为目标版本基线
   - 删除 `openspec/specs/` 下非版本目录的 capability 文件夹（仅保留 `vX.Y.Z/`）

   **归档顺序**：若 `openspec archive` 因 MODIFIED 目标不存在而失败，先归档仅含 ADDED 新 capability 的 change，再重试依赖项。REMOVED 头不匹配时修正 delta 或手工合并后重试。

   若脚本失败，按上述步骤手工完成并重试。

5. **更新项目基线引用**

   将 `openspec/project.md` 中 OpenSpec 基线参考由旧版本（如 `v2.0.2`）更新为目标版本（如 `v2.0.3`），包括 `openspec/specs/vX.Y.Z/spec.md` 路径与相关评审检查项。

6. **校验与摘要**

   运行 `openspec validate --strict`（若有活跃 change 应已为空）。
   确认 `openspec list` 返回空 changes。
   确认 `openspec/specs/<target>/spec.md` 已生成且目录含新 capability。

**Output On Success**

```
## 收版完成

**目标版本:** v2.0.3
**源基线:** v2.0.2
**合并规格:** openspec/specs/v2.0.3/spec.md
**已归档 change:** <列表>
**Specs:** ✓ 已合并至版本基线（无 archive 目录）

project.md 基线引用已更新为 v2.0.3。
```

**Output On Success With Warnings**

```
## 收版完成（含警告）

**目标版本:** v2.0.3
**Warnings:**
- ci-acr-github-secrets 含 6 项未完成任务仍已收版
- history-event-unit-denorm 含 2 项未完成任务仍已收版
```

**Guardrails**

- 目标版本名必须由用户提供或在本轮对话中明确；不要自动选版本
- **禁止**保留 `openspec/changes/archive/` 内容；archive 后立即删除
- **禁止**在收版后保留 `openspec/specs/<capability>/` 散落目录（仅保留 `vX.Y.Z/` 基线）
- MODIFIED/REMOVED delta 要求 capability spec 已存在 — 必须先 split 源基线
- 收版后必须更新 `openspec/project.md` 基线版本引用

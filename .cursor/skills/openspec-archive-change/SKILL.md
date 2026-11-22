---
name: openspec-archive-change
description: Archive a completed change in the experimental workflow. Use when the user wants to finalize and archive a change after implementation is complete.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.0"
---

Archive a completed change in the experimental workflow.

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show only active changes (not already archived).
   Include the schema used for each change if available.

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Check artifact completion status**

   Run `openspec status --change "<name>" --json` to check artifact completion.

   Parse the JSON to understand:
   - `schemaName`: The workflow being used
   - `artifacts`: List of artifacts with their status (`done` or other)

   **If any artifacts are not `done`:**
   - Display warning listing incomplete artifacts
   - Use **AskUserQuestion tool** to confirm user wants to proceed
   - Proceed if user confirms

3. **Check task completion status**

   Read the tasks file (typically `tasks.md`) to check for incomplete tasks.

   Count tasks marked with `- [ ]` (incomplete) vs `- [x]` (complete).

   **If incomplete tasks found:**
   - Display warning showing count of incomplete tasks
   - Use **AskUserQuestion tool** to confirm user wants to proceed
   - Proceed if user confirms

   **If no tasks file exists:** Proceed without task-related warning.

4. **Assess delta spec sync state**

   Check for delta specs at `openspec/changes/<name>/specs/`. If none exist, proceed without sync prompt.

   **If delta specs exist:**
   - Compare each delta with `openspec/specs/<domain>/spec.md`
   - Show a combined summary before prompting

   **Prompt options:**
   - If changes needed: "Sync now (recommended)", "Skip sync (mains already updated)"
   - If already synced: "Delete change now", "Sync anyway", "Cancel"

   If user chooses sync, merge deltas into main specs (agent-driven or `openspec archive <name> -y` then remove any archive folder the CLI created). If skip, use `openspec archive <name> -y --skip-specs` only as a merge helper when useful—**either way this repo deletes the change afterward**.

5. **Perform the archive（本仓库：合并后删除，不留 archive 历史）**

   1. 合并 delta 进主规格，或确认主规格已含变更后选择 skip-specs。
   2. **删除** `openspec/changes/<name>/`（`Remove-Item -Recurse -Force` / `rm -rf`）。
   3. 若 CLI 曾写入 `openspec/changes/archive/YYYY-MM-DD-<name>/`，**一并删除**。
   4. 不要创建或保留 dated archive 作为默认结果。

6. **Display summary**

```
## Archive Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Change directory:** deleted
**Specs:** ✓ Synced to main specs (or "Skip sync - mains already updated")

All artifacts complete. All tasks complete.
```

**Guardrails**
- Always prompt for change selection if not provided
- Use artifact graph (openspec status --json) for completion checking
- Don't block archive on warnings - just inform and confirm
- **Do not** keep `openspec/changes/archive/` history by default (solo OPC)
- Show clear summary of what happened
- If delta specs exist, always show the combined summary before prompting

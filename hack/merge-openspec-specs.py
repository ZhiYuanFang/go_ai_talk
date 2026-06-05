#!/usr/bin/env python3
"""合并 openspec/specs 下全部 capability spec.md 为单一版本基线文件。"""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SPECS_DIR = ROOT / "openspec" / "specs"
OUT = SPECS_DIR / "v1.0.3" / "spec.md"


def main() -> None:
    files = sorted(SPECS_DIR.glob("*/spec.md"))
    # 排除已生成的 v1.0.3（若重复运行）
    files = [f for f in files if f.parent.name != "v1.0.3"]
    if not files:
        raise SystemExit("no spec files found")

    parts: list[str] = []
    parts.append("# OpenSpec 基线规格 v1.0.3\n\n")
    parts.append(
        "> 本文件由 `openspec/specs` 下全部 capability 规格于 **v1.0.3** 合并而成，"
        "作为该版本的确定性规则基线，便于按版本查阅。\n\n"
    )
    parts.append(
        "> 后续新变更请基于本文件对照增量，或通过 OpenSpec 新建 change 再合并至下一版本规格。\n\n"
    )
    parts.append("## 目录\n\n")
    for f in files:
        cap = f.parent.name
        parts.append(f"- [{cap}](#{cap})\n")
    parts.append("\n---\n\n")

    for f in files:
        cap = f.parent.name
        content = f.read_text(encoding="utf-8").strip()
        parts.append(f"## {cap}\n\n")
        parts.append(f"<!-- source: openspec/specs/{cap}/spec.md -->\n\n")
        parts.append(content)
        parts.append("\n\n---\n\n")

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text("".join(parts), encoding="utf-8", newline="\n")
    text = OUT.read_text(encoding="utf-8")
    print(f"merged {len(files)} specs -> {OUT.relative_to(ROOT)}")
    print(f"lines={len(text.splitlines())} bytes={OUT.stat().st_size}")


if __name__ == "__main__":
    main()

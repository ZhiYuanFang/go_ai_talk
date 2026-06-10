#!/usr/bin/env python3
"""从合并基线 spec.md 拆出 openspec/specs/<capability>/spec.md。"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SPECS_DIR = ROOT / "openspec" / "specs"


def split_baseline(baseline_path: Path) -> dict[str, str]:
    text = baseline_path.read_text(encoding="utf-8")
    # 跳过头部与目录，从第一个 capability 章节开始
    parts = re.split(r"\n---\n\n", text)
    caps: dict[str, str] = {}
    for part in parts:
        m = re.match(r"## ([a-z0-9-]+)\n\n<!-- source: [^>]+ -->\n\n(.*)", part, re.DOTALL)
        if not m:
            continue
        cap, body = m.group(1), m.group(2).strip()
        caps[cap] = body + "\n"
    return caps


def main() -> None:
    version = sys.argv[1] if len(sys.argv) > 1 else "v1.0.3"
    baseline = SPECS_DIR / version / "spec.md"
    if not baseline.is_file():
        raise SystemExit(f"baseline not found: {baseline}")

    caps = split_baseline(baseline)
    if not caps:
        raise SystemExit("no capabilities parsed")

    for cap, content in sorted(caps.items()):
        out = SPECS_DIR / cap / "spec.md"
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(content, encoding="utf-8", newline="\n")

    print(f"split {len(caps)} capabilities from {baseline.relative_to(ROOT)}")


if __name__ == "__main__":
    main()

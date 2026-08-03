#!/usr/bin/env python3
"""将活跃 OpenSpec 变更收进目标版本基线，不保留 archive 目录。

流程：split 源基线 → openspec archive 各 change → 删除 archive → merge 目标版本 → 清理 capability 目录。
"""
from __future__ import annotations

import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SPECS_DIR = ROOT / "openspec" / "specs"
CHANGES_DIR = ROOT / "openspec" / "changes"
ARCHIVE_DIR = CHANGES_DIR / "archive"
VERSION_RE = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def parse_version(name: str) -> tuple[int, int, int]:
    m = VERSION_RE.match(name)
    if not m:
        raise SystemExit(f"invalid version: {name!r} (expected vX.Y.Z)")
    return int(m.group(1)), int(m.group(2)), int(m.group(3))


def list_version_dirs() -> list[str]:
    return sorted(
        (p.name for p in SPECS_DIR.iterdir() if p.is_dir() and VERSION_RE.match(p.name)),
        key=parse_version,
    )


def resolve_source_version(target: str, explicit: str | None) -> str:
    if explicit:
        return explicit
    versions = list_version_dirs()
    if target not in versions:
        prior = [v for v in versions if parse_version(v) < parse_version(target)]
        if not prior:
            raise SystemExit(f"no source baseline before {target}")
        return prior[-1]
    # 目标已存在：以次新版本为源重新合并
    prior = [v for v in versions if v != target and parse_version(v) < parse_version(target)]
    if not prior:
        raise SystemExit(f"cannot infer source baseline for existing {target}")
    return prior[-1]


def resolve_openspec_cmd() -> list[str]:
    """Windows 上 `openspec` 常为 .cmd/.ps1，CreateProcess 直接找无扩展名会失败。"""
    which = shutil.which("openspec")
    if which:
        return [which]
    # 常见 npm global 路径回退
    for name in ("openspec.cmd", "openspec"):
        for base in (
            Path(os.environ.get("APPDATA", "")) / "npm",
            Path(r"D:\Program Files\nodejs\node_global"),
            Path(r"C:\Program Files\nodejs"),
        ):
            cand = base / name
            if cand.is_file():
                return [str(cand)]
    return ["openspec"]


def run(cmd: list[str], *, check: bool = True) -> subprocess.CompletedProcess[str]:
    if cmd and cmd[0] == "openspec":
        cmd = resolve_openspec_cmd() + cmd[1:]
    print("$", " ".join(cmd))
    return subprocess.run(cmd, cwd=ROOT, text=True, check=check)


def active_changes() -> list[str]:
    if not CHANGES_DIR.is_dir():
        return []
    return sorted(
        p.name
        for p in CHANGES_DIR.iterdir()
        if p.is_dir() and p.name != "archive" and not p.name.startswith(".")
    )


def remove_archive_for(change: str) -> None:
    if not ARCHIVE_DIR.is_dir():
        return
    for p in sorted(ARCHIVE_DIR.iterdir(), reverse=True):
        if p.is_dir() and p.name.endswith(f"-{change}"):
            shutil.rmtree(p)
            print(f"removed archive dir: {p.name}")
            return


def cleanup_capability_dirs() -> int:
    removed = 0
    for p in SPECS_DIR.iterdir():
        if p.is_dir() and not VERSION_RE.match(p.name):
            shutil.rmtree(p)
            removed += 1
    return removed


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit("usage: archive-openspec-to-version.py <target-version> [source-version]")

    target = sys.argv[1]
    parse_version(target)
    source = resolve_source_version(target, sys.argv[2] if len(sys.argv) > 2 else None)

    print(f"source baseline: {source}")
    print(f"target version:  {target}")

    run([sys.executable, "hack/split-openspec-baseline.py", source])

    changes = active_changes()
    if not changes:
        print("no active changes; merging capability specs only")
    else:
        print(f"archiving {len(changes)} change(s): {', '.join(changes)}")
        for change in changes:
            print(f"\n=== {change} ===")
            result = run(["openspec", "archive", change, "--yes"], check=False)
            if result.returncode != 0:
                raise SystemExit(f"openspec archive failed for {change}")
            remove_archive_for(change)

    run([sys.executable, "hack/merge-openspec-specs.py", target])
    n = cleanup_capability_dirs()
    print(f"cleaned {n} capability spec directories")
    print(f"done -> openspec/specs/{target}/spec.md")


if __name__ == "__main__":
    main()

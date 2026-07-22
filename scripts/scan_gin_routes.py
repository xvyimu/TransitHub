#!/usr/bin/env python3
"""Scan router/*.go for Gin route registrations (static, best-effort)."""
from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ROUTER = ROOT / "router"

PAT_METHOD = re.compile(
    r"""\.(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|Any)\(\s*[\"']([^\"']+)[\"']"""
)
PAT_HANDLE = re.compile(
    r"""method:\s*http\.Method(\w+)\s*,\s*path:\s*[\"']([^\"']+)[\"']"""
)


def scan_file(path: Path) -> list[dict]:
    text = path.read_text(encoding="utf-8")
    items: list[dict] = []
    for m in PAT_METHOD.finditer(text):
        items.append(
            {
                "method": m.group(1).upper() if m.group(1) != "Any" else "ANY",
                "path_fragment": m.group(2),
                "src": path.name,
            }
        )
    for m in PAT_HANDLE.finditer(text):
        items.append(
            {
                "method": m.group(1).upper(),
                "path_fragment": m.group(2),
                "src": f"{path.name}:permission_table",
            }
        )
    return items


def main() -> None:
    out: dict[str, list] = {}
    total = 0
    for f in sorted(ROUTER.glob("*.go")):
        if f.name.endswith("_test.go"):
            continue
        items = scan_file(f)
        out[f.name] = items
        total += len(items)
        print(f"{f.name}: {len(items)}")
    dest_dir = ROOT / "docs" / "gateway"
    dest_dir.mkdir(parents=True, exist_ok=True)
    raw = dest_dir / "_route_scan_raw.json"
    raw.write_text(json.dumps(out, indent=2, ensure_ascii=False), encoding="utf-8")
    print(f"total={total} wrote {raw.relative_to(ROOT)}")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Read-only diff helpers for OpenAPI vs known Gin surfaces (Phase1 WP-G G5).

Does not modify business code. Exit codes:
  0 — ran successfully (drift may still be reported)
  2 — missing inputs
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def load_ops(openapi_path: Path) -> set[str]:
    doc = json.loads(openapi_path.read_text(encoding="utf-8"))
    ops: set[str] = set()
    for path, methods in doc.get("paths", {}).items():
        for method, _body in methods.items():
            if method in ("get", "post", "put", "delete", "patch", "head", "options"):
                ops.add(f"{method.upper()} {path}")
    return ops


def norm_param(s: str) -> str:
    s = re.sub(r"\{[^}]+\}", "{var}", s)
    s = re.sub(r":[A-Za-z0-9_]+", "{var}", s)
    s = s.replace("/*path", "/{var}")
    return s


# Curated Gin relay+video paths (keep in sync with docs/gateway/ROUTE_TABLE.md §C)
GIN_RELAY_OPS = {
    "GET /v1/models",
    "GET /v1/models/{var}",
    "GET /v1beta/models",
    "GET /v1beta/openai/models",
    "POST /pg/chat/completions",
    "GET /v1/realtime",
    "POST /v1/messages",
    "POST /v1/completions",
    "POST /v1/chat/completions",
    "POST /v1/responses",
    "POST /v1/responses/compact",
    "POST /v1/edits",
    "POST /v1/images/generations",
    "POST /v1/images/edits",
    "POST /v1/embeddings",
    "POST /v1/audio/transcriptions",
    "POST /v1/audio/translations",
    "POST /v1/audio/speech",
    "POST /v1/rerank",
    "POST /v1/engines/{var}/embeddings",
    "POST /v1/models/{var}",
    "POST /v1/moderations",
    "POST /suno/submit/{var}",
    "POST /suno/fetch",
    "GET /suno/fetch/{var}",
    "POST /v1beta/models/{var}",
    "GET /v1/videos/{var}/content",
    "POST /v1/video/generations",
    "GET /v1/video/generations/{var}",
    "POST /v1/videos/{var}/remix",
    "POST /v1/videos",
    "GET /v1/videos/{var}",
    "POST /kling/v1/videos/text2video",
    "POST /kling/v1/videos/image2video",
    "GET /kling/v1/videos/text2video/{var}",
    "GET /kling/v1/videos/image2video/{var}",
    "POST /jimeng/",
}


CONSOLE_SUBSET = {
    "GET /healthz",
    "GET /livez",
    "GET /readyz",
    "GET /api/status",
    "POST /api/user/login",
    "GET /api/user/logout",
    "GET /api/user/self",
}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--api",
        type=Path,
        default=ROOT / "docs" / "openapi" / "api.json",
    )
    parser.add_argument(
        "--relay",
        type=Path,
        default=ROOT / "docs" / "openapi" / "relay.json",
    )
    args = parser.parse_args()
    if not args.api.is_file() or not args.relay.is_file():
        print("missing openapi json", file=sys.stderr)
        return 2

    api_ops = {norm_param(x) for x in load_ops(args.api)}
    relay_ops = {norm_param(x) for x in load_ops(args.relay)}

    print("=== Console subset vs api.json (+probes not expected in api.json) ===")
    for op in sorted(CONSOLE_SUBSET):
        if op.startswith("GET /health") or op.startswith("GET /live") or op.startswith("GET /ready"):
            print(f"  {op}: probe (document in console-subset only)")
            continue
        hit = norm_param(op) in api_ops
        print(f"  {op}: {'OK' if hit else 'MISSING in api.json'}")

    print("\n=== Relay curated Gin vs relay.json (normalized) ===")
    gin = {norm_param(x) for x in GIN_RELAY_OPS}
    only_gin = sorted(gin - relay_ops)
    only_doc = sorted(relay_ops - gin)
    print(f"gin_curated={len(gin)} openapi={len(relay_ops)}")
    print(f"in gin not openapi ({len(only_gin)}):")
    for x in only_gin:
        print(f"  {x}")
    print(f"in openapi not gin curated ({len(only_doc)}):")
    for x in only_doc[:40]:
        print(f"  {x}")
    if len(only_doc) > 40:
        print(f"  ... {len(only_doc) - 40} more")

    print("\n=== api.json scale ===")
    print(f"api.json ops (normalized)={len(api_ops)}")
    print("Full management Gin inventory is fragment-based; see scan_gin_routes.py + OPENAPI_AUDIT.md")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

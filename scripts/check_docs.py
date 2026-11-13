#!/usr/bin/env python3
"""Check local Markdown links, Swagger assets, and local OpenAPI references."""

import hashlib
import json
import re
from pathlib import Path

root = Path(__file__).resolve().parents[1]
docs = root / "docs"
for file in [root / "README.md", *docs.rglob("*.md")]:
    for target in re.findall(r"\[[^\]]*\]\(([^)]+)\)", file.read_text()):
        if target.startswith(("https://", "http://", "#", "mailto:")):
            continue
        path = target.split("#", 1)[0]
        assert (file.parent / path).exists(), f"Broken link: {file.relative_to(root)} -> {target}"
manifest = json.loads((docs / "api/assets/manifest.json").read_text())
for name, digest in manifest["sha256"].items():
    content = (docs / "api/assets" / name).read_bytes()
    assert hashlib.sha256(content).hexdigest() == digest, f"Swagger asset checksum mismatch: {name}"
spec = json.loads((docs / "api/openapi.json").read_text())


def check_refs(value):
    if isinstance(value, dict):
        if "$ref" in value:
            pointer = value["$ref"]
            assert pointer.startswith("#/"), f"External reference: {pointer}"
            resolved = spec
            for part in pointer[2:].split("/"):
                resolved = resolved[part.replace("~1", "/").replace("~0", "~")]
        for child in value.values():
            check_refs(child)
    elif isinstance(value, list):
        for child in value:
            check_refs(child)


check_refs(spec)
print(
    f"PASS: local documentation links, OpenAPI references, and Swagger UI {manifest['version']} checksums."
)

#!/usr/bin/env python3
"""Synchronize application version metadata from the root VERSION file."""

import argparse
import json
import re
from pathlib import Path

NUMBER = r"(?:0|[1-9][0-9]*)"
IDENTIFIER = rf"(?:{NUMBER}|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)"
SEMVER = re.compile(rf"(?P<base>{NUMBER}\.{NUMBER}\.{NUMBER})(?:-{IDENTIFIER}(?:\.{IDENTIFIER})*)?")
ROOT = Path(__file__).resolve().parents[1]


def validate(version):
    if not SEMVER.fullmatch(version) or len(version) > 127:
        raise ValueError(
            "Expected MAJOR.MINOR.PATCH with an optional prerelease, without a v prefix"
        )
    return version


def read_version(root=ROOT):
    return validate((root / "VERSION").read_text().strip())


def synchronize(root=ROOT, version=None, write=False, tag=None):
    version = validate(version) if version is not None else read_version(root)
    if tag is not None and tag != f"v{version}":
        raise ValueError(f"Release tag must match VERSION: v{version}")
    changes = []

    def update(relative, mutate):
        path = root / relative
        original = json.loads(path.read_text())
        updated = json.loads(path.read_text())
        mutate(updated)
        if updated != original:
            changes.append(relative)
            if write:
                # Preserve the OpenAPI formatter's existing layout.
                if relative == "docs/api/openapi.json":
                    text = re.sub(
                        r'("version"\s*:\s*)"[^"]+"',
                        rf'\g<1>"{version}"',
                        path.read_text(),
                        count=1,
                    )
                else:
                    text = json.dumps(updated, indent=2) + "\n"
                path.write_text(text)

    for folder in ("", "frontend/", "mobile/"):
        update(folder + "package.json", lambda data: data.update(version=version))

        def lock(data):
            data["version"] = version
            data["packages"][""]["version"] = version

        update(folder + "package-lock.json", lock)

    def expo(data):
        data["expo"]["version"] = SEMVER.fullmatch(version)["base"]
        data["expo"].setdefault("extra", {})["releaseVersion"] = version

    update("mobile/app.json", expo)
    update("docs/api/openapi.json", lambda data: data["info"].update(version=version))
    backend = root / "backend/cmd/server/main.go"
    source = backend.read_text()
    updated, count = re.subn(r'(\bVersion\s*=\s*)"[^"]+"', rf'\g<1>"{version}"', source, count=1)
    if count != 1:
        raise ValueError("Backend version variable was not found")
    if source != updated:
        changes.append("backend/cmd/server/main.go")
        if write:
            backend.write_text(updated)
    if write:
        (root / "VERSION").write_text(version + "\n")
    elif changes:
        raise ValueError(
            "Version drift: " + ", ".join(changes) + "; run make version-set VERSION=" + version
        )
    return version


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("check").add_argument("--tag")
    commands.add_parser("set").add_argument("version")
    args = parser.parse_args()
    if args.command == "set":
        version = synchronize(version=args.version, write=True)
    else:
        version = synchronize(tag=args.tag)
    print(version)


if __name__ == "__main__":
    main()

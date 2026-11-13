#!/usr/bin/env python3
"""Validate release inputs and assemble digest-pinned image manifests."""

import json
import os
import re
import sys
from pathlib import Path
from urllib.parse import urlsplit

from project_version import synchronize, validate

SERVICES = ("backend", "frontend", "mobile")


def metadata(repository, ref, api_url=""):
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repository):
        raise ValueError("Expected an owner/repository name")
    if ref.startswith("refs/tags/"):
        version = ref.removeprefix("refs/tags/")
        if not version.startswith("v"):
            raise ValueError("Release tags must be vMAJOR.MINOR.PATCH with an optional prerelease")
        validate(version[1:])
        parsed = urlsplit(api_url)
        if (
            parsed.scheme not in ("https", "http")
            or not parsed.hostname
            or parsed.hostname.endswith(".invalid")
            or parsed.username is not None
            or parsed.password is not None
            or parsed.query
            or parsed.fragment
            or any(c.isspace() for c in api_url)
        ):
            raise ValueError("Set repository variable EXPO_PUBLIC_API_URL to the deployed API URL")
    elif match := re.fullmatch(r"refs/pull/([1-9][0-9]*)/merge", ref):
        version = "pr-" + match[1]
    else:
        raise ValueError("Only pull requests and version tags are supported")
    return {
        "version": version,
        "image_prefix": "ghcr.io/" + repository.lower(),
        "prerelease": str(version.startswith("v") and "-" in version).lower(),
    }


def manifest(directory, prefix, version, revision):
    images = {}
    for service in SERVICES:
        record = json.loads((directory / f"{service}.json").read_text())
        image = f"{prefix}/{service}"
        if record["image"] != image or not re.fullmatch(r"sha256:[0-9a-f]{64}", record["digest"]):
            raise ValueError(f"Invalid image record: {service}")
        images[service] = {"tag": f"{image}:{version}", "digest": f"{image}@{record['digest']}"}
    return {
        "version": version,
        "app_version": version.removeprefix("v"),
        "revision": revision,
        "platform": "linux/amd64",
        "images": images,
    }


def main():
    if sys.argv[1] == "prepare":
        app_version = synchronize(
            tag=os.environ["GITHUB_REF"].removeprefix("refs/tags/")
            if os.environ["GITHUB_REF"].startswith("refs/tags/")
            else None
        )
        result = metadata(
            os.environ["GITHUB_REPOSITORY"],
            os.environ["GITHUB_REF"],
            os.getenv("EXPO_PUBLIC_API_URL", ""),
        )
        result["app_version"] = app_version
        result["matrix"] = json.dumps(
            {
                "include": [
                    {"service": service, "context": "." if service == "backend" else f"./{service}"}
                    for service in SERVICES
                ]
            }
        )
        with Path(os.environ["GITHUB_OUTPUT"]).open("a") as output:
            for key, value in result.items():
                output.write(f"{key}={value}\n")
    elif sys.argv[1] == "manifest":
        result = manifest(
            Path("digests"),
            os.environ["IMAGE_PREFIX"],
            os.environ["RELEASE_TAG"],
            os.environ["GITHUB_SHA"],
        )
        destination = Path("release")
        destination.mkdir(exist_ok=True)
        (destination / "images.json").write_text(json.dumps(result, indent=2) + "\n")
        (destination / "images.env").write_text(
            "".join(
                f"{name.upper()}_IMAGE={image['digest']}\n"
                for name, image in result["images"].items()
            )
        )
        lines = [
            f"Release {result['version']}",
            "",
            f"Commit: `{result['revision']}`",
            "",
            "Production images (linux/amd64):",
            "",
        ]
        lines.extend(f"- `{image['digest']}`" for image in result["images"].values())
        lines.extend(
            [
                "",
                "Download images.env for digest-pinned Compose deployments. The mobile image exports JavaScript bundles; it is not a native application runtime.",
                "",
            ]
        )
        (destination / "images.md").write_text("\n".join(lines))
    else:
        raise ValueError("Expected prepare or manifest")


if __name__ == "__main__":
    main()

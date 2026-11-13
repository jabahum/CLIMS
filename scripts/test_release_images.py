import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from release_images import SERVICES, manifest, metadata
from project_version import ROOT, read_version


class ReleaseImagesTest(unittest.TestCase):
    def test_prepare_outputs_only_applications_at_shared_version(self):
        with tempfile.TemporaryDirectory() as folder:
            output = Path(folder) / "outputs"
            version = read_version()
            subprocess.run(
                [sys.executable, str(ROOT / "scripts/release_images.py"), "prepare"],
                check=True,
                env={
                    **os.environ,
                    "GITHUB_REPOSITORY": "Owner/CLIMS",
                    "GITHUB_REF": f"refs/tags/v{version}",
                    "GITHUB_OUTPUT": str(output),
                    "EXPO_PUBLIC_API_URL": "https://api.example.org/api/v1",
                },
            )
            values = dict(line.split("=", 1) for line in output.read_text().splitlines())
            self.assertEqual(values["app_version"], version)
            self.assertEqual(values["version"], "v" + version)
            matrix = json.loads(values["matrix"])["include"]
            self.assertEqual(
                {item["service"] for item in matrix}, {"backend", "frontend", "mobile"}
            )

    def test_stable_release_and_registry_case(self):
        result = metadata("mohuganda/CLIMS", "refs/tags/v1.2.3", "https://api.example.org/api/v1")
        self.assertEqual(
            result,
            {"version": "v1.2.3", "image_prefix": "ghcr.io/mohuganda/clims", "prerelease": "false"},
        )

    def test_prerelease(self):
        result = metadata("owner/repo", "refs/tags/v0.1.0-rc.1", "https://api.example.org")
        self.assertEqual(result["prerelease"], "true")

    def test_pull_request_does_not_need_production_url(self):
        self.assertEqual(metadata("owner/repo", "refs/pull/12/merge")["version"], "pr-12")

    def test_rejects_invalid_version_tags(self):
        for version in ("v1", "v1.2", "v01.2.3", "v1.2.3-01", "v1.2.3+build.1", "v1.2.3\n"):
            with self.subTest(version=version), self.assertRaises(ValueError):
                metadata("owner/repo", "refs/tags/" + version, "https://api.example.org")

    def test_rejects_missing_or_invalid_release_urls(self):
        for url in (
            "",
            "api.example.org",
            "https://example.invalid/api/v1",
            "https://user:password@example.org",
            "https://example.org/#fragment",
            "https://example.org\n",
        ):
            with self.subTest(url=url), self.assertRaises(ValueError):
                metadata("owner/repo", "refs/tags/v1.2.3", url)

    def test_rejects_branch_builds_and_output_injection(self):
        with self.assertRaises(ValueError):
            metadata("owner/repo", "refs/heads/main")
        with self.assertRaises(ValueError):
            metadata("owner/repo\nother=bad", "refs/pull/1/merge")

    def test_manifest_pins_every_service_and_rejects_bad_digests(self):
        with tempfile.TemporaryDirectory() as folder:
            directory = Path(folder)
            for service in SERVICES:
                (directory / f"{service}.json").write_text(
                    json.dumps(
                        {"image": f"ghcr.io/owner/repo/{service}", "digest": "sha256:" + "a" * 64}
                    )
                )
            result = manifest(directory, "ghcr.io/owner/repo", "v1.2.3", "b" * 40)
            self.assertEqual(set(result["images"]), {"backend", "frontend", "mobile"})
            self.assertEqual(result["app_version"], "1.2.3")
            self.assertEqual(
                result["images"]["backend"]["digest"],
                "ghcr.io/owner/repo/backend@sha256:" + "a" * 64,
            )
            (directory / "backend.json").write_text(
                json.dumps({"image": "ghcr.io/owner/repo/backend", "digest": "invalid"})
            )
            with self.assertRaises(ValueError):
                manifest(directory, "ghcr.io/owner/repo", "v1.2.3", "b" * 40)

    def test_manifest_requires_all_services(self):
        with tempfile.TemporaryDirectory() as folder, self.assertRaises(FileNotFoundError):
            manifest(Path(folder), "ghcr.io/owner/repo", "v1.2.3", "b" * 40)


if __name__ == "__main__":
    unittest.main()

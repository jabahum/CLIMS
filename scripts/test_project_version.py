import json
import shutil
import tempfile
import unittest
from pathlib import Path

from project_version import ROOT, read_version, synchronize


class ProjectVersionTest(unittest.TestCase):
    def fixture(self, folder):
        root = Path(folder)
        files = [
            "VERSION",
            "mobile/app.json",
            "docs/api/openapi.json",
            "backend/cmd/server/main.go",
        ]
        for prefix in ("", "frontend/", "mobile/"):
            files.extend([prefix + "package.json", prefix + "package-lock.json"])
        for relative in files:
            destination = root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / relative, destination)
        return root

    def test_updates_clients_locks_backend_and_api_together(self):
        with tempfile.TemporaryDirectory() as folder:
            root = self.fixture(folder)
            synchronize(root, "2.3.4-rc.2", write=True)
            self.assertEqual(synchronize(root, tag="v2.3.4-rc.2"), "2.3.4-rc.2")
            self.assertEqual(read_version(root), "2.3.4-rc.2")
            for prefix in ("", "frontend/", "mobile/"):
                self.assertEqual(
                    json.loads((root / (prefix + "package.json")).read_text())["version"],
                    "2.3.4-rc.2",
                )
                lock = json.loads((root / (prefix + "package-lock.json")).read_text())
                self.assertEqual(lock["version"], "2.3.4-rc.2")
                self.assertEqual(lock["packages"][""]["version"], "2.3.4-rc.2")
            expo = json.loads((root / "mobile/app.json").read_text())["expo"]
            self.assertEqual(expo["version"], "2.3.4")
            self.assertEqual(expo["extra"]["releaseVersion"], "2.3.4-rc.2")
            self.assertEqual(
                json.loads((root / "docs/api/openapi.json").read_text())["info"]["version"],
                "2.3.4-rc.2",
            )
            self.assertIn('"2.3.4-rc.2"', (root / "backend/cmd/server/main.go").read_text())

    def test_detects_lockfile_drift_without_changing_files(self):
        with tempfile.TemporaryDirectory() as folder:
            root = self.fixture(folder)
            path = root / "frontend/package-lock.json"
            lock = json.loads(path.read_text())
            lock["packages"][""]["version"] = "9.9.9"
            path.write_text(json.dumps(lock))
            before = path.read_bytes()
            with self.assertRaisesRegex(ValueError, "frontend/package-lock.json"):
                synchronize(root)
            self.assertEqual(path.read_bytes(), before)

    def test_rejects_tag_mismatch_and_invalid_version_before_writing(self):
        with tempfile.TemporaryDirectory() as folder:
            root = self.fixture(folder)
            before = (root / "VERSION").read_bytes()
            with self.assertRaises(ValueError):
                synchronize(root, tag="v99.99.99")
            with self.assertRaises(ValueError):
                synchronize(root, "v1.2.3", write=True)
            self.assertEqual((root / "VERSION").read_bytes(), before)


if __name__ == "__main__":
    unittest.main()

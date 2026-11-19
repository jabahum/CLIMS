"""Exercise server deployment success and failure without contacting Docker or SSH."""

import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


class DeploymentTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name).resolve()
        self.release = self.root / "releases/new"
        (self.release / "scripts").mkdir(parents=True)
        self.script = self.release / "scripts/deploy_release.sh"
        shutil.copyfile(Path(__file__).with_name("deploy_release.sh"), self.script)
        (self.root / "shared").mkdir()
        (self.root / "shared/prod.env").write_text("DB_PASSWORD=test\nJWT_SECRET=test\n")
        (self.release / "images.env").write_text("BACKEND_IMAGE=test@sha256:abc\n")
        self.old = self.root / "releases/old"
        self.old.mkdir()
        (self.root / "current").symlink_to(self.old)
        binaries = self.root / "bin"
        binaries.mkdir()
        docker = binaries / "docker"
        docker.write_text(
            '#!/bin/bash\nprintf "%s\\n" "$*" >> "$DEPLOY_TEST_LOG"\n'
            'if [[ " $* " == *" up "* && ${FAIL_HEALTH:-} == 1 ]]; then exit 1; fi\n'
        )
        docker.chmod(0o755)
        # Lock behavior is provided by util-linux on the production Linux host.
        flock = binaries / "flock"
        flock.write_text("#!/bin/bash\nexit 0\n")
        flock.chmod(0o755)
        self.log = self.root / "docker.log"
        self.environment = {
            **os.environ,
            "PATH": f"{binaries}:{os.environ['PATH']}",
            "DEPLOY_TEST_LOG": str(self.log),
        }

    def run_deployment(self, **environment):
        return subprocess.run(
            ["bash", str(self.script), str(self.root)],
            env={**self.environment, **environment},
            capture_output=True,
            text=True,
        )

    def test_success_preserves_previous_and_uses_health_checks(self):
        result = self.run_deployment()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.root / "current").resolve(), self.release)
        self.assertEqual((self.root / "previous").resolve(), self.old)
        commands = self.log.read_text().splitlines()
        self.assertEqual(len(commands), 4)
        self.assertIn("--project-name clims-prod", commands[-1])
        self.assertIn("--no-build --wait --wait-timeout 300", commands[-1])

    def test_failed_health_does_not_mark_release_successful(self):
        result = self.run_deployment(FAIL_HEALTH="1")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual((self.root / "current").resolve(), self.old)
        self.assertFalse((self.root / "previous").exists())

    def test_missing_server_secrets_fails_before_docker(self):
        (self.root / "shared/prod.env").unlink()
        self.assertNotEqual(self.run_deployment().returncode, 0)
        self.assertFalse(self.log.exists())

    def test_recovery_through_symlink_uses_physical_release_path(self):
        self.assertEqual(self.run_deployment().returncode, 0)
        (self.old / "scripts").mkdir()
        shutil.copyfile(self.script, self.old / "scripts/deploy_release.sh")
        shutil.copyfile(self.release / "images.env", self.old / "images.env")
        result = subprocess.run(
            ["bash", str(self.root / "previous/scripts/deploy_release.sh"), str(self.root)],
            env=self.environment,
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.root / "current").resolve(), self.old)
        self.assertEqual((self.root / "previous").resolve(), self.release)


if __name__ == "__main__":
    unittest.main()

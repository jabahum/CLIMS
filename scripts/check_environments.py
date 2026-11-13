#!/usr/bin/env python3
"""Verify the actual merged Compose models without exposing environment secrets."""

import json
import os
import subprocess
import tempfile
from pathlib import Path

root = Path(__file__).resolve().parents[1]
base = root / "infra/compose/docker-compose.yml"
with tempfile.NamedTemporaryFile(mode="w", suffix=".env") as environment:
    environment.write(
        "DB_PASSWORD=validation-only\nJWT_SECRET=validation-only\nEXPO_PUBLIC_API_URL=https://example.invalid/api/v1\n"
    )
    environment.flush()
    models = {}
    for mode in ("dev", "prod"):
        result = subprocess.run(
            [
                "docker",
                "compose",
                "--env-file",
                environment.name,
                "-f",
                str(base),
                "-f",
                str(base.with_name(f"docker-compose.{mode}.yml")),
                "--profile",
                "artifacts",
                "config",
                "--format",
                "json",
            ],
            check=True,
            capture_output=True,
            text=True,
        )
        model = json.loads(result.stdout)
        models[mode] = model
        services = model["services"]
        assert set(services) == {"backend", "frontend", "mobile", "postgres", "nginx"}
        for name, service in services.items():
            build = service["build"]
            assert build["dockerfile"].endswith(f"Dockerfile.{mode}"), (mode, name)
            assert (Path(build["context"]) / build["dockerfile"]).is_file(), (mode, name)
            assert service["image"].endswith(":" + mode), (mode, name)
        if mode == "dev":
            assert model["name"] == "clims", "Development volume namespace changed"
            assert len(services["nginx"]["ports"]) == 2
            assert services["postgres"].get("ports")
            for name in ("backend", "frontend", "mobile"):
                assert any(m["type"] == "bind" for m in services[name]["volumes"])
            assert "mobile" in services["nginx"]["depends_on"]
        else:
            assert model["name"] == "clims-prod"
            assert len(services["nginx"]["ports"]) == 1
            for name in ("backend", "frontend", "postgres"):
                assert not services[name].get("ports"), name
            for name in ("backend", "frontend", "postgres", "nginx"):
                assert not any(m["type"] == "bind" for m in services[name].get("volumes", [])), name
            assert services["mobile"]["profiles"] == ["artifacts"]
            assert "mobile" not in services["nginx"]["depends_on"]
    pinned = {
        f"{service.upper()}_IMAGE": f"ghcr.io/example/clims/{service}@sha256:" + "a" * 64
        for service in ("backend", "frontend", "mobile")
    }
    result = subprocess.run(
        [
            "docker",
            "compose",
            "--env-file",
            environment.name,
            "-f",
            str(base),
            "-f",
            str(base.with_name("docker-compose.prod.yml")),
            "--profile",
            "artifacts",
            "config",
            "--format",
            "json",
        ],
        check=True,
        capture_output=True,
        text=True,
        env={**os.environ, **pinned},
    )
    released = json.loads(result.stdout)
    for service, definition in released["services"].items():
        expected = pinned.get(
            f"{service.upper()}_IMAGE", models["prod"]["services"][service]["image"]
        )
        assert definition["image"] == expected, service
    assert (
        models["dev"]["volumes"]["postgres_data"]["name"]
        != models["prod"]["volumes"]["postgres_data"]["name"]
    )
print(
    "PASS: dev/prod Dockerfile parity, development mounts, production isolation, gateway ports, separate database volumes, and release image overrides."
)

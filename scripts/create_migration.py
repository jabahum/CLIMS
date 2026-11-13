#!/usr/bin/env python3
"""Create a sequential SQL migration pair with the project's pinned migrate CLI."""

import argparse
import os
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def create_migration(name, directory=None):
    if not name or not re.fullmatch(r"[a-z][a-z0-9]*(?:_[a-z0-9]+)*", name):
        raise ValueError(
            "Use a lowercase snake_case name, e.g. make migrate-create NAME=add_stock_index"
        )
    backend = ROOT / "backend"
    match = re.search(
        r"^\s*github\.com/golang-migrate/migrate/v4\s+(v[0-9]+\.[0-9]+\.[0-9]+)\s*$",
        (backend / "go.mod").read_text(),
        re.MULTILINE,
    )
    if not match:
        raise ValueError("Cannot find the pinned golang-migrate version in backend/go.mod")
    destination = directory if directory is not None else backend / "internal/database/migrations"
    if not destination.is_dir():
        raise ValueError("The migrations directory must already exist")
    subprocess.run(
        [
            "go",
            "run",
            "-tags",
            "postgres",
            f"github.com/golang-migrate/migrate/v4/cmd/migrate@{match[1]}",
            "create",
            "-ext",
            "sql",
            "-dir",
            str(destination),
            "-seq",
            "-digits",
            "6",
            name,
        ],
        cwd=backend,
        check=True,
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("name", nargs="?", default=os.getenv("MIGRATION_NAME"))
    args = parser.parse_args()
    try:
        create_migration(args.name)
    except (ValueError, FileNotFoundError) as error:
        parser.error(str(error))
    except subprocess.CalledProcessError as error:
        raise SystemExit(error.returncode) from error


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Read-only smoke check of the Nginx web/API and Metro gateways."""

import json
import os
from urllib.error import HTTPError
from urllib.request import Request, urlopen

base = os.getenv("STACK_URL", "http://localhost:8080").rstrip("/")
metro = os.getenv("METRO_URL", "http://localhost:19000").rstrip("/")


def get(url, headers=None):
    with urlopen(Request(url, headers=headers or {}), timeout=30) as response:
        assert response.status == 200, f"Unexpected HTTP status: {url}"
        return response.read()


html = get(base + "/").decode()
assert "CLIMS" in html and "A connected foundation" in html, "Next.js page was not served"
assert json.loads(get(base + "/health/ready"))["data"]["database"] == "connected"
assert "swagger-ui" in get(base + "/docs/").decode()
assert json.loads(get(base + "/docs/openapi.json"))["openapi"] == "3.0.3"
try:
    get(base + "/api/v1/auth/me")
except HTTPError as error:
    assert error.code == 401, "API route did not reach JWT middleware"
else:
    raise AssertionError("Profile route accepted a request without a token")
environment = os.getenv("STACK_ENV", "dev")
if environment not in ("dev", "prod"):
    raise ValueError("STACK_ENV must be dev or prod")
if environment == "dev":
    assert "packager-status:running" in get(metro + "/status").decode()
    manifest = json.loads(
        get(metro + "/", {"Accept": "application/expo+json", "Expo-Platform": "android"})
    )
    assert "launchAsset" in manifest, "Expo did not return a native application manifest"
print(
    f"PASS: {environment} Nginx gateway serves Next.js, API readiness, Swagger, and authenticated API protection."
    + (" Expo Metro manifest is available." if environment == "dev" else "")
)

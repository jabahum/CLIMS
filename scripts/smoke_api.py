#!/usr/bin/env python3
"""Check the running API against a disposable database (creates one test account)."""

import json
import sys
import uuid
from urllib.error import HTTPError
from urllib.request import Request, urlopen

base = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"


def request(path, status, data=None, token=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = Request(base + path, data=json.dumps(data).encode() if data else None, headers=headers)
    try:
        result = urlopen(req, timeout=10)
    except HTTPError as err:
        result = err
    with result:
        assert result.code == status, f"{path}: expected {status}, got {result.code}"
        return json.load(result)


request("/health/live", 200)
request("/health/ready", 200)
suffix = uuid.uuid4().hex[:12]
account = {
    "username": "smoke_" + suffix,
    "email": suffix + "@example.test",
    "password": "smoke-password-123",
}
user_id = request("/api/v1/auth/register", 201, account)["data"]["user_id"]
request("/api/v1/auth/register", 409, account)
request("/api/v1/auth/login", 401, {"email": account["email"], "password": "incorrect"})
token = request(
    "/api/v1/auth/login", 200, {"email": account["email"], "password": account["password"]}
)["data"]["token"]
profile = request("/api/v1/auth/me", 200, token=token)["data"]
assert profile["id"] == user_id and profile["email"] == account["email"]
request("/api/v1/auth/me", 401)
print(
    "PASS: health, registration, duplicate email, login, incorrect password, authenticated profile, and missing token."
)

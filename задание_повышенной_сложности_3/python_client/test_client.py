"""
Тесты HTTP-клиента AuthClient.
HTTP-вызовы перехватываются библиотекой responses (без реального сервера).
"""
import time
from datetime import datetime, timedelta, timezone

import jwt
import pytest
import responses as resp_mock

from client import AuthClient
from jwt_verifier import SECRET, ALGORITHM

BASE_URL = "http://localhost:8080"


def make_token(username: str = "admin") -> str:
    payload = {
        "username": username,
        "sub": username,
        "iat": datetime.now(tz=timezone.utc),
        "exp": datetime.now(tz=timezone.utc) + timedelta(hours=1),
    }
    return jwt.encode(payload, SECRET, algorithm=ALGORITHM)


# ---------------------------------------------------------------------------
# login
# ---------------------------------------------------------------------------

class TestLogin:
    @resp_mock.activate
    def test_login_stores_token(self):
        token = make_token()
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", json={"token": token})

        c = AuthClient(BASE_URL)
        result = c.login("admin", "password123")

        assert result == token
        assert c.token == token

    @resp_mock.activate
    def test_login_sends_credentials(self):
        token = make_token()
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", json={"token": token})

        c = AuthClient(BASE_URL)
        c.login("admin", "password123")

        body = resp_mock.calls[0].request.body
        assert b"admin" in body
        assert b"password123" in body

    @resp_mock.activate
    def test_login_401_raises(self):
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", status=401,
                      json={"error": "invalid credentials"})
        c = AuthClient(BASE_URL)
        with pytest.raises(Exception):
            c.login("admin", "wrong")


# ---------------------------------------------------------------------------
# authenticated requests
# ---------------------------------------------------------------------------

class TestAuthenticatedRequests:
    @resp_mock.activate
    def test_get_profile_sends_bearer(self):
        token = make_token("user")
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", json={"token": token})
        resp_mock.add(resp_mock.GET, f"{BASE_URL}/api/profile",
                      json={"username": "user", "role": "user"})

        c = AuthClient(BASE_URL)
        c.login("user", "secret")
        profile = c.get_profile()

        assert profile["username"] == "user"
        auth_header = resp_mock.calls[1].request.headers["Authorization"]
        assert auth_header == f"Bearer {token}"

    @resp_mock.activate
    def test_get_items_returns_list(self):
        token = make_token()
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", json={"token": token})
        resp_mock.add(resp_mock.GET, f"{BASE_URL}/api/items",
                      json=[{"id": 1, "name": "A", "owner": "admin"}])

        c = AuthClient(BASE_URL)
        c.login("admin", "password123")
        items = c.get_items()

        assert isinstance(items, list)
        assert items[0]["name"] == "A"

    @resp_mock.activate
    def test_create_item(self):
        token = make_token()
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/auth/login", json={"token": token})
        resp_mock.add(resp_mock.POST, f"{BASE_URL}/api/items",
                      json={"id": 1, "name": "Widget", "owner": "admin"}, status=201)

        c = AuthClient(BASE_URL)
        c.login("admin", "password123")
        item = c.create_item("Widget")

        assert item["name"] == "Widget"

    def test_unauthenticated_raises_value_error(self):
        c = AuthClient(BASE_URL)
        with pytest.raises(ValueError, match="not authenticated"):
            c.get_profile()

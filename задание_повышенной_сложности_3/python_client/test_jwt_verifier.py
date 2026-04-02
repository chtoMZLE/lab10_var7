"""
Тесты верификации JWT-токенов (имитируем токены Go-сервиса).

Используем PyJWT для генерации токенов с тем же секретом и алгоритмом,
что и Go-сервис — это доказывает совместимость форматов.
"""
import time
from datetime import datetime, timedelta, timezone

import jwt
import pytest

from jwt_verifier import (
    SECRET,
    ALGORITHM,
    verify_token,
    get_username,
    is_token_valid,
)


def make_go_like_token(
    username: str = "admin",
    exp_delta: timedelta = timedelta(hours=24),
) -> str:
    """Создаёт токен в формате Go-сервиса (поля username + sub)."""
    now = datetime.now(tz=timezone.utc)
    payload = {
        "username": username,
        "sub": username,
        "iat": now,
        "exp": now + exp_delta,
    }
    return jwt.encode(payload, SECRET, algorithm=ALGORITHM)


# ---------------------------------------------------------------------------
# verify_token
# ---------------------------------------------------------------------------

class TestVerifyToken:
    def test_valid_token_returns_payload(self):
        token = make_go_like_token("alice")
        payload = verify_token(token)
        assert payload["username"] == "alice"
        assert payload["sub"] == "alice"

    def test_expired_token_raises(self):
        token = make_go_like_token(exp_delta=timedelta(seconds=-1))
        with pytest.raises(jwt.ExpiredSignatureError):
            verify_token(token)

    def test_wrong_secret_raises(self):
        token = jwt.encode({"sub": "hacker", "exp": time.time() + 3600}, "wrong-key", algorithm="HS256")
        with pytest.raises(jwt.InvalidSignatureError):
            verify_token(token)

    def test_malformed_token_raises(self):
        with pytest.raises(jwt.DecodeError):
            verify_token("not.a.valid.jwt")

    def test_empty_token_raises(self):
        with pytest.raises(jwt.DecodeError):
            verify_token("")

    def test_algorithm_none_rejected(self):
        # alg=none — критическая уязвимость, должна отклоняться
        import base64, json
        header = base64.urlsafe_b64encode(
            json.dumps({"alg": "none", "typ": "JWT"}).encode()
        ).rstrip(b"=").decode()
        payload_b = base64.urlsafe_b64encode(
            json.dumps({"sub": "attacker", "exp": int(time.time()) + 9999}).encode()
        ).rstrip(b"=").decode()
        fake_token = f"{header}.{payload_b}."
        with pytest.raises(jwt.exceptions.PyJWTError):
            verify_token(fake_token)


# ---------------------------------------------------------------------------
# get_username
# ---------------------------------------------------------------------------

class TestGetUsername:
    def test_extracts_username_field(self):
        token = make_go_like_token("bob")
        assert get_username(token) == "bob"

    def test_falls_back_to_sub(self):
        payload = {"sub": "carol", "exp": datetime.now(tz=timezone.utc) + timedelta(hours=1)}
        token = jwt.encode(payload, SECRET, algorithm=ALGORITHM)
        assert get_username(token) == "carol"


# ---------------------------------------------------------------------------
# is_token_valid
# ---------------------------------------------------------------------------

class TestIsTokenValid:
    def test_valid_token_returns_true(self):
        assert is_token_valid(make_go_like_token()) is True

    def test_expired_token_returns_false(self):
        token = make_go_like_token(exp_delta=timedelta(seconds=-1))
        assert is_token_valid(token) is False

    def test_bad_signature_returns_false(self):
        token = jwt.encode({"sub": "x", "exp": time.time() + 3600}, "bad", algorithm="HS256")
        assert is_token_valid(token) is False

    def test_garbage_returns_false(self):
        assert is_token_valid("garbage") is False

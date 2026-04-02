"""HTTP-клиент Go-сервиса с JWT-аутентификацией."""
import requests
from jwt_verifier import verify_token


class AuthClient:
    def __init__(self, base_url: str):
        self._base_url = base_url.rstrip("/")
        self._token: str | None = None

    @property
    def token(self) -> str | None:
        return self._token

    def login(self, username: str, password: str) -> str:
        """Авторизуется и сохраняет JWT. Возвращает токен."""
        resp = requests.post(
            f"{self._base_url}/auth/login",
            json={"username": username, "password": password},
            timeout=10,
        )
        resp.raise_for_status()
        self._token = resp.json()["token"]
        return self._token

    def get_profile(self) -> dict:
        return self._get("/api/profile")

    def get_items(self) -> list:
        return self._get("/api/items")

    def create_item(self, name: str) -> dict:
        return self._post("/api/items", {"name": name})

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _auth_headers(self) -> dict:
        if not self._token:
            raise ValueError("not authenticated — call login() first")
        return {"Authorization": f"Bearer {self._token}"}

    def _get(self, path: str) -> dict | list:
        resp = requests.get(
            f"{self._base_url}{path}",
            headers=self._auth_headers(),
            timeout=10,
        )
        resp.raise_for_status()
        return resp.json()

    def _post(self, path: str, body: dict) -> dict:
        resp = requests.post(
            f"{self._base_url}{path}",
            json=body,
            headers=self._auth_headers(),
            timeout=10,
        )
        resp.raise_for_status()
        return resp.json()

"""Юнит-тесты FastAPI-сервиса (graceful shutdown + эндпоинты)."""
import pytest
from fastapi.testclient import TestClient

from main import app, reset_db


@pytest.fixture(autouse=True)
def clean_db():
    """Перед каждым тестом сбрасываем состояние БД."""
    reset_db()
    yield
    reset_db()


# TestClient сам запускает lifespan (startup + shutdown)
client = TestClient(app)


class TestLifespan:
    def test_app_ready_after_startup(self):
        with TestClient(app) as c:
            assert app.state.ready is True
            resp = c.get("/health")
            assert resp.status_code == 200

    def test_app_not_ready_after_shutdown(self):
        with TestClient(app):
            pass
        assert app.state.ready is False


class TestHealth:
    def test_health_returns_ok(self):
        resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json() == {"status": "ok"}


class TestItems:
    def test_list_items_empty(self):
        resp = client.get("/items")
        assert resp.status_code == 200
        assert resp.json() == []

    def test_create_item(self):
        payload = {"name": "Widget", "description": "A test widget", "price": 9.99}
        resp = client.post("/items", json=payload)
        assert resp.status_code == 201
        data = resp.json()
        assert data["id"] == 1
        assert data["name"] == "Widget"
        assert data["price"] == 9.99

    def test_create_item_increments_id(self):
        client.post("/items", json={"name": "A", "price": 1.0})
        resp = client.post("/items", json={"name": "B", "price": 2.0})
        assert resp.json()["id"] == 2

    def test_list_items_after_create(self):
        client.post("/items", json={"name": "X", "price": 5.0})
        resp = client.get("/items")
        assert resp.status_code == 200
        assert len(resp.json()) == 1

    def test_get_item_by_id(self):
        client.post("/items", json={"name": "Gadget", "price": 19.99})
        resp = client.get("/items/1")
        assert resp.status_code == 200
        assert resp.json()["name"] == "Gadget"

    def test_get_item_not_found(self):
        resp = client.get("/items/999")
        assert resp.status_code == 404
        assert resp.json()["detail"] == "item not found"

    def test_create_item_without_description(self):
        resp = client.post("/items", json={"name": "NoDesc", "price": 3.0})
        assert resp.status_code == 201
        assert resp.json()["description"] is None

    def test_create_item_missing_name(self):
        resp = client.post("/items", json={"price": 5.0})
        assert resp.status_code == 422

    def test_create_item_missing_price(self):
        resp = client.post("/items", json={"name": "NoPriceItem"})
        assert resp.status_code == 422

    def test_create_item_invalid_price_type(self):
        resp = client.post("/items", json={"name": "Bad", "price": "not-a-number"})
        assert resp.status_code == 422

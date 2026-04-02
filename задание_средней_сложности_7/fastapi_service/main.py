"""FastAPI-сервис с graceful shutdown через lifespan."""
from contextlib import asynccontextmanager
from typing import List

from fastapi import FastAPI, HTTPException

from models import Item, ItemCreate

# In-memory хранилище
_items: List[Item] = []
_next_id: int = 1


def get_db() -> dict:
    """Возвращает ссылки на in-memory хранилище (для тестов)."""
    return {"items": _items, "next_id": _next_id}


def reset_db() -> None:
    """Сброс состояния — используется в тестах."""
    global _items, _next_id
    _items = []
    _next_id = 1


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Управляет жизненным циклом приложения (startup / graceful shutdown)."""
    # --- startup ---
    print("Service starting up...")
    app.state.ready = True
    yield
    # --- shutdown ---
    print("Service shutting down gracefully...")
    app.state.ready = False
    reset_db()
    print("Cleanup done.")


app = FastAPI(title="Items API", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/items", response_model=List[Item])
def list_items():
    return _items


@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: int):
    for item in _items:
        if item.id == item_id:
            return item
    raise HTTPException(status_code=404, detail="item not found")


@app.post("/items", response_model=Item, status_code=201)
def create_item(payload: ItemCreate):
    global _next_id
    item = Item(id=_next_id, **payload.model_dump())
    _next_id += 1
    _items.append(item)
    return item

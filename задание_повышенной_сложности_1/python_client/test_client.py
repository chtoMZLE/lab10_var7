"""
Юнит-тесты Python gRPC-клиента.

Поднимаем Python gRPC-сервер в отдельном потоке через тот же .proto,
чтобы тестировать клиента без зависимости от Go.
"""
import sys
import os
import threading
import time
from concurrent import futures

import grpc
import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "gen"))

from gen import user_pb2, user_pb2_grpc
from client import UserServiceClient


# ---------------------------------------------------------------------------
# Вспомогательный Python-сервер (тот же интерфейс, что и Go)
# ---------------------------------------------------------------------------

class _UserServiceServicer(user_pb2_grpc.UserServiceServicer):
    def __init__(self):
        self._users = {}
        self._next_id = 1
        self._lock = threading.Lock()

    def CreateUser(self, request, context):
        if not request.name:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "name is required")
        if not request.email:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "email is required")
        with self._lock:
            u = user_pb2.User(id=self._next_id, name=request.name, email=request.email)
            self._users[self._next_id] = u
            self._next_id += 1
        return u

    def GetUser(self, request, context):
        with self._lock:
            u = self._users.get(request.id)
        if u is None:
            context.abort(grpc.StatusCode.NOT_FOUND, f"user {request.id} not found")
        return u

    def ListUsers(self, request, context):
        with self._lock:
            users = list(self._users.values())
        return user_pb2.ListUsersResponse(users=users)


@pytest.fixture(scope="module")
def grpc_server():
    """Запускает in-process gRPC сервер для тестов."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    user_pb2_grpc.add_UserServiceServicer_to_server(_UserServiceServicer(), server)
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    yield f"127.0.0.1:{port}"
    server.stop(grace=1)


@pytest.fixture()
def client(grpc_server):
    c = UserServiceClient(grpc_server)
    yield c
    c.close()


# ---------------------------------------------------------------------------
# Тесты
# ---------------------------------------------------------------------------

class TestCreateUser:
    def test_create_returns_user(self, client):
        u = client.create_user("Alice", "alice@example.com")
        assert u.name == "Alice"
        assert u.email == "alice@example.com"
        assert u.id > 0

    def test_create_increments_id(self, client):
        u1 = client.create_user("Bob", "bob@example.com")
        u2 = client.create_user("Carol", "carol@example.com")
        assert u2.id == u1.id + 1

    def test_create_missing_name_raises(self, client):
        with pytest.raises(grpc.RpcError) as exc:
            client.create_user("", "a@b.com")
        assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT

    def test_create_missing_email_raises(self, client):
        with pytest.raises(grpc.RpcError) as exc:
            client.create_user("Dave", "")
        assert exc.value.code() == grpc.StatusCode.INVALID_ARGUMENT


class TestGetUser:
    def test_get_existing_user(self, client):
        created = client.create_user("Eve", "eve@example.com")
        got = client.get_user(created.id)
        assert got.name == "Eve"
        assert got.email == "eve@example.com"

    def test_get_not_found_raises(self, client):
        with pytest.raises(grpc.RpcError) as exc:
            client.get_user(99999)
        assert exc.value.code() == grpc.StatusCode.NOT_FOUND


class TestListUsers:
    def test_list_returns_created_users(self, client):
        before = len(client.list_users())
        client.create_user("Frank", "f@f.com")
        client.create_user("Grace", "g@g.com")
        after = client.list_users()
        assert len(after) == before + 2

"""gRPC-клиент для UserService."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "gen"))

import grpc
from gen import user_pb2, user_pb2_grpc


class UserServiceClient:
    def __init__(self, address: str = "localhost:50051"):
        self._channel = grpc.insecure_channel(address)
        self._stub = user_pb2_grpc.UserServiceStub(self._channel)

    def close(self):
        self._channel.close()

    def create_user(self, name: str, email: str):
        req = user_pb2.CreateUserRequest(name=name, email=email)
        return self._stub.CreateUser(req)

    def get_user(self, user_id: int):
        req = user_pb2.GetUserRequest(id=user_id)
        return self._stub.GetUser(req)

    def list_users(self):
        req = user_pb2.ListUsersRequest()
        return self._stub.ListUsers(req).users

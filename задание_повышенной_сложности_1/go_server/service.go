package main

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "lab10/task_advanced1/gen/userpb"
)

// UserServiceServer реализует pb.UserServiceServer.
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	mu     sync.Mutex
	users  map[int32]*pb.User
	nextID int32
}

func NewUserServiceServer() *UserServiceServer {
	return &UserServiceServer{
		users:  make(map[int32]*pb.User),
		nextID: 1,
	}
}

func (s *UserServiceServer) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	u := &pb.User{Id: s.nextID, Name: req.Name, Email: req.Email}
	s.users[s.nextID] = u
	s.nextID++
	return u, nil
}

func (s *UserServiceServer) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("user %d not found", req.Id))
	}
	return u, nil
}

func (s *UserServiceServer) ListUsers(_ context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]*pb.User, 0, len(s.users))
	for _, u := range s.users {
		list = append(list, u)
	}
	return &pb.ListUsersResponse{Users: list}, nil
}

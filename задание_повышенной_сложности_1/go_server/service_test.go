package main

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "lab10/task_advanced1/gen/userpb"
)

const bufSize = 1024 * 1024

// newTestClient поднимает gRPC-сервер через bufconn (in-memory) и возвращает клиента.
func newTestClient(t *testing.T) (pb.UserServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterUserServiceServer(srv, NewUserServiceServer())

	go func() {
		srv.Serve(lis) //nolint:errcheck
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}
	return pb.NewUserServiceClient(conn), cleanup
}

func TestCreateUser(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	resp, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), resp.Id)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "alice@example.com", resp.Email)
}

func TestCreateUser_MissingName(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	_, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{Email: "a@b.com"})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateUser_MissingEmail(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	_, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{Name: "Bob"})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetUser(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	created, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{
		Name: "Carol", Email: "carol@example.com",
	})
	require.NoError(t, err)

	got, err := client.GetUser(context.Background(), &pb.GetUserRequest{Id: created.Id})
	require.NoError(t, err)
	assert.Equal(t, "Carol", got.Name)
}

func TestGetUser_NotFound(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	_, err := client.GetUser(context.Background(), &pb.GetUserRequest{Id: 999})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListUsers_Empty(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	resp, err := client.ListUsers(context.Background(), &pb.ListUsersRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Users, 0)
}

func TestListUsers_AfterCreate(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	client.CreateUser(context.Background(), &pb.CreateUserRequest{Name: "D", Email: "d@d.com"})    //nolint:errcheck
	client.CreateUser(context.Background(), &pb.CreateUserRequest{Name: "E", Email: "e@e.com"})    //nolint:errcheck

	resp, err := client.ListUsers(context.Background(), &pb.ListUsersRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Users, 2)
}

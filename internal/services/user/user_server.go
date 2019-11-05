package user

import (
	"context"
	pb "do/api/user"

	"google.golang.org/grpc"
)

type server struct{}

func RegisterServer(s *grpc.Server) {
	pb.RegisterUserServer(s, &server{})
}

func (s *server) GetUsers(ctx context.Context, in *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	var users []*pb.UserResponse
	users = append(users, &pb.UserResponse{
		Id:   "test-id1",
		Name: "test-name",
	})

	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}

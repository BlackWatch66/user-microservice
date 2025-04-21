package grpc

import (
	"context"

	pb "github.com/blackwatch66/user-microservice/api/grpc/proto"
	"github.com/blackwatch66/user-microservice/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer 实现了 UserServiceServer gRPC 接口
type UserServer struct {
	pb.UnimplementedUserServiceServer // 嵌入未实现的 server 以确保向前兼容
	userService service.UserService
}

// NewUserServer 创建一个新的 UserServer
func NewUserServer(userService service.UserService) *UserServer {
	return &UserServer{userService: userService}
}

// CreateUser 实现 gRPC 的 CreateUser 方法
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Email and password are required")
	}

	// 注意：这里直接调用了 Register，它内部会处理密码哈希
	user, err := s.userService.Register(ctx, req.Email, req.Password)
	if err != nil {
		// 根据 service 层返回的错误转换 gRPC 状态码
		if err.Error() == "email already exists" {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to create user: %v", err)
	}

	return &pb.CreateUserResponse{
		UserId: uint64(user.ID),
		Email:  user.Email,
	}, nil
}

// ValidateToken 实现 gRPC 的 ValidateToken 方法
func (s *UserServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Token is required")
	}

	claims, err := s.userService.ValidateToken(ctx, req.Token)
	if err != nil {
		// Token 无效或过期
		return &pb.ValidateTokenResponse{Valid: false}, nil // 返回无效，不暴露具体错误给调用方
        // 或者根据错误类型返回更具体的 gRPC 错误码
        // return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: uint64(claims.UserID),
		Email:  claims.Email,
	}, nil
} 
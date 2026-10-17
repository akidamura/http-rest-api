package grpcserver

import (
	"fmt"
	"log/slog"
	"net"

	userv1 "github.com/akidamura/http-rest-api/internal/pb/user/v1"
	"google.golang.org/grpc"
)

// Server реализует gRPC-сервис UserService
type Server struct {
	userv1.UnimplementedUserServiceServer // встроили - обязательно
	userRepo                              UserRepository
}

// New создает и возвращает новый экземпляр gRPC-сервера
func New(userRepo UserRepository) *Server {
	return &Server{
		userRepo: userRepo,
	}
}

// Start запускает gRPC-сервер на указанном адресе.
// Не блокирует вызывающую горутину
func (s *Server) Start(addr string, logger *slog.Logger) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, s)

	go func() {
		logger.Info("Starting gRPC server", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "err", err)
		}
	}()

	return nil
}

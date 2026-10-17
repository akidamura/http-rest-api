package grpcserver

import (
	"context"
	"errors"

	"github.com/akidamura/http-rest-api/internal/app/model"
	userv1 "github.com/akidamura/http-rest-api/internal/pb/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateUser обрабатывает запрос на регистрацию нового пользователя.
// Валидирует входные данные и возвращает ошибку InvalidArgument при их некорректности
func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user := &model.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, model.ErrEmailRequired) ||
			errors.Is(err, model.ErrPasswordRequired) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &userv1.CreateUserResponse{User: userToProto(createdUser)}, nil
}

// GetUser возвращает данные пользователя по его ID.
// Возвращает NotFound, если пользователь с таким ID не существует
func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	id := int(req.GetId())
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	switch {
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	case user == nil:
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userv1.GetUserResponse{User: userToProto(user)}, nil
}

// UpdateUser обновляет данные существующего пользователя.
// Возвращает NotFound, если пользователь не найден
func (s *Server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	id := int(req.GetId())
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user := &model.User{
		ID:       id,
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	updatedUser, err := s.userRepo.Update(ctx, user)
	switch {
	case err != nil:
		// 1. Ошибка БД или валидации
		if errors.Is(err, model.ErrEmailRequired) || errors.Is(err, model.ErrPasswordRequired) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	case updatedUser == nil:
		// 2. Ошибки нет, но и юзера нет - значит не найден
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userv1.UpdateUserResponse{User: userToProto(updatedUser)}, nil
}

// DeleteUser удаляет пользователя по ID.
// Если пользователя нет, возвращает успешный ответ (идемпотентность)
func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	id := int(req.GetId())
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &userv1.DeleteUserResponse{}, nil
}

package grpcserver

import (
	"time"

	"github.com/akidamura/http-rest-api/internal/app/model"
	userv1 "github.com/akidamura/http-rest-api/internal/pb/user/v1"
)

// userToProto конвертирует доменную модель в protobuf-сообщение.
// Отбрасывает чувствительные данные (пароли, хеши), которые не должны покидать сервер
func userToProto(u *model.User) *userv1.User {
	if u == nil {
		return nil
	}

	var createdAt string
	if u.CreatedAt != nil {
		createdAt = u.CreatedAt.Format(time.RFC3339)
	}

	return &userv1.User{
		Id:        int64(u.ID),
		Email:     u.Email,
		CreatedAt: createdAt,
	}
}

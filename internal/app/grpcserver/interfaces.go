package grpcserver

import (
	"context"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// UserRepository определяет контракт хранилища, необходимый для работы gRPC-сервиса
type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	Update(ctx context.Context, u *model.User) (*model.User, error)
	Delete(ctx context.Context, id int) error
}

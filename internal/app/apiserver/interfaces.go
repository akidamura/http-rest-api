package apiserver

import (
	"context"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// UserRepository — то, что нужно хендлерам от хранилища пользователей.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
}

// Pinger — то, что нужно /health, чтобы проверить соединение.
type Pinger interface {
	Ping(ctx context.Context) error
}

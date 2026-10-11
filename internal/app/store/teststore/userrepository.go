package teststore

import (
	"context"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

type UserRepository struct {
	store  *Store
	users  map[int]*model.User
	nextID int
}

func (r *UserRepository) Create(_ context.Context, u *model.User) (*model.User, error) {
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}
	r.nextID++
	u.ID = r.nextID
	r.users[u.ID] = u
	return u, nil
}

func (r *UserRepository) FindByID(_ context.Context, id int) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (r *UserRepository) Update(_ context.Context, u *model.User) (*model.User, error) {
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}
	if _, ok := r.users[u.ID]; !ok {
		return nil, nil
	}
	r.users[u.ID] = u
	return u, nil
}

func (r *UserRepository) Delete(_ context.Context, id int) error {
	delete(r.users, id)
	return nil
}

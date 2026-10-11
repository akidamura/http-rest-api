package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// UserRepository ...
type UserRepository struct {
	store *Store
}

// Create ...
func (r *UserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	query := `INSERT INTO users (email, encrypted_password) VALUES (?, ?)`
	res, err := r.store.db.ExecContext(ctx, query, u.Email, u.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.ID = int(id)

	return u, nil
}

// FindByEmail ...
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, email, encrypted_password, created_at
              FROM users WHERE email = ?`

	err := r.store.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

// FindByID ...
func (r *UserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, email, encrypted_password, created_at
			  FROM users WHERE id = ?`

	err := r.store.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) (*model.User, error) {
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	query := `UPDATE users SET email = ?, encrypted_password = ? WHERE id = ?`
	_, err := r.store.db.ExecContext(ctx, query, u.Email, u.EncryptedPassword, u.ID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.store.db.ExecContext(ctx, query, id)
	return err
}

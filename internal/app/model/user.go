package model

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
)

// User ...
type User struct {
	ID                int        `json:"id"`
	Email             string     `json:"email"`
	Password          string     `json:"-"`
	EncryptedPassword string     `json:"-"`
	CreatedAt         *time.Time `json:"created_at,omitempty"` // '*' потому что omitempty не работает с time.Time, структура, а не примитив
}

// BeforeCreate валидирует пользователя и хеширует пароль
func (u *User) BeforeCreate() error {
	// Очистить открытый пароль
	defer func() {
		u.Password = ""
	}()

	if err := u.validate(); err != nil {
		return err
	}

	return u.hashPassword()
}

func (u *User) validate() error {
	u.Email = strings.TrimSpace(u.Email)

	if u.Email == "" {
		return ErrEmailRequired
	}
	if u.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

// hashPassword хеширует открытый пароль и сохраняет результат
func (u *User) hashPassword() error {
	hash, err := encryptString(u.Password)
	if err != nil {
		return err
	}
	u.EncryptedPassword = hash
	return nil
}

func encryptString(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

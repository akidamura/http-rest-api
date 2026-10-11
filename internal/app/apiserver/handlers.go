package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akidamura/http-rest-api/internal/app/model"
)

// handleHello ...
func (s *APIServer) handleHello() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}

// handleCreateUser ...
func (s *APIServer) handleCreateUser() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Декодируем JSON
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// 2. Собираем модель
		user := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}

		// 3. Сохраняем (валидация + хеш внутри BeforeCrate)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		user, err := s.store.User().Create(ctx, user)
		if err != nil {
			// Ошибки валидации → 400
			if errors.Is(err, model.ErrEmailRequired) ||
				errors.Is(err, model.ErrPasswordRequired) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// 4. Ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// handleGetUser ...
func (s *APIServer) handleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Достать id из url
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		// 2. Найти в БД
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// s — это получатель метода (func (s *APIServer) handleCreateUser).
		// Замыкание захватывает s. Поэтому внутри func(w, r) можно писать s.store
		user, err := s.store.User().FindByID(ctx, id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// 3. Собираем json для ответа
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

package apiserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyUser
)

// authenticateUser проверяет куку с session и кладёт пользователя в контекст.
func (s *APIServer) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		// кладём user в контекст
		ctxWithUser := context.WithValue(r.Context(), ctxKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}

// setRequestID добавляет X-Request-ID в заголовок ответа и в контекст запроса.
func (s *APIServer) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Add("X-Request-ID", id)

		ctx := context.WithValue(r.Context(), ctxKeyRequestID, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// logRequest логирует каждый входящий запрос:
// метод, путь, статус, время выполнения.
func (s *APIServer) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем w, чтобы захватить статус
		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)

		// Выбираем уровень лога по статусу
		var level slog.Level
		switch {
		case rw.code >= 500:
			level = slog.LevelError
		case rw.code >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
		}

		s.logger.LogAttrs(r.Context(), level, "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.code),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", requestIDFromCtx(r.Context())))
	})
}

// requestIDFromCtx достаёт request_id из контекста.
func requestIDFromCtx(ctx context.Context) string {
	id, ok := ctx.Value(ctxKeyRequestID).(string)
	if !ok {
		return ""
	}
	return id
}

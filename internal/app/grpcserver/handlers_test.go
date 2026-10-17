package grpcserver

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/model"
	userv1 "github.com/akidamura/http-rest-api/internal/pb/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// --- Фейк репозитория (как в apiserver, но для пакета grpcserver) ---

type fakeUserRepo struct {
	createFn   func(ctx context.Context, u *model.User) (*model.User, error)
	findByIDFn func(ctx context.Context, id int) (*model.User, error)
	updateFn   func(ctx context.Context, u *model.User) (*model.User, error)
	deleteFn   func(ctx context.Context, id int) error
}

func (f *fakeUserRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	return f.createFn(ctx, u)
}
func (f *fakeUserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	return f.findByIDFn(ctx, id)
}
func (f *fakeUserRepo) Update(ctx context.Context, u *model.User) (*model.User, error) {
	return f.updateFn(ctx, u)
}
func (f *fakeUserRepo) Delete(ctx context.Context, id int) error {
	return f.deleteFn(ctx, id)
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		createFn: func(_ context.Context, u *model.User) (*model.User, error) {
			u.ID = 42
			return u, nil
		},
		findByIDFn: func(_ context.Context, id int) (*model.User, error) {
			return &model.User{ID: id, Email: "test@example.org"}, nil
		},
		updateFn: func(_ context.Context, u *model.User) (*model.User, error) {
			return u, nil
		},
		deleteFn: func(_ context.Context, id int) error {
			return nil
		},
	}
}

// --- Инфраструктура тестов (bufconn) ---

// newTestClient создает изолированный gRPC-сервер в памяти (bufconn)
// и возвращает клиент для тестов.
func newTestClient(t *testing.T, repo UserRepository) userv1.UserServiceClient {
	t.Helper() // Помечаем функцию как хелпер, чтобы при падении показывалась строка теста, а не эта функция

	// 1. Создаем виртуальный TCP-листенер
	lis := bufconn.Listen(1024 * 1024)

	// 2. Создаем наш gRPC-сервер
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, New(repo))

	// Запускаем сервер в фоне
	go func() {
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	// Создаем клиента
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	// t.Cleanup выполнится автоматически, когда тест закончится.
	// Это гарантирует, что мы не оставим висящих горутин и соединений.
	t.Cleanup(func() {
		srv.Stop()
		conn.Close()
	})

	return userv1.NewUserServiceClient(conn)
}

// --- Тесты ---

func TestCreateUser_OK(t *testing.T) {
	repo := newFakeUserRepo()
	client := newTestClient(t, repo)

	req := &userv1.CreateUserRequest{
		Email:    "test@example.org",
		Password: "secret123",
	}

	resp, err := client.CreateUser(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.GetUser().GetEmail() != req.Email {
		t.Errorf("expected email %q, got %q", req.Email, resp.GetUser().GetEmail())
	}
	if resp.GetUser().GetId() <= 0 {
		t.Errorf("expected ID > 0, got %d", resp.GetUser().GetId())
	}
}

func TestCreateUser_ValidationError(t *testing.T) {
	// Настраиваем фейк так, чтобы он вернул ошибку валидации
	// (как будто репозиторий вызвал BeforeCreate и она упала)
	repo := newFakeUserRepo()
	repo.createFn = func(_ context.Context, _ *model.User) (*model.User, error) {
		return nil, model.ErrEmailRequired
	}
	client := newTestClient(t, repo)

	req := &userv1.CreateUserRequest{Email: "", Password: "secret123"}
	_, err := client.CreateUser(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Распаковываем ошибку в gRPC-статус
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code InvalidArgument, got %v", st.Code())
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	repo.findByIDFn = func(_ context.Context, _ int) (*model.User, error) {
		return nil, nil // Имитируем "не найден"
	}
	client := newTestClient(t, repo)

	req := &userv1.GetUserRequest{Id: 999}
	_, err := client.GetUser(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Распаковываем ошибку в gRPC-статус
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %v", st.Code())
	}
}

package auth

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/model"
)

type fakeUserRepository struct {
	createParams CreateUserParams
	createErr    error
	findUser     model.User
	findErr      error
}

func (r *fakeUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (model.User, error) {
	r.createParams = params
	if r.createErr != nil {
		return model.User{}, r.createErr
	}

	return model.User{
		ID:           10001,
		Phone:        params.Phone,
		PasswordHash: params.PasswordHash,
		Nickname:     params.Nickname,
	}, nil
}

func (r *fakeUserRepository) FindUserByPhone(ctx context.Context, phone string) (model.User, error) {
	if r.findErr != nil {
		return model.User{}, r.findErr
	}
	return r.findUser, nil
}

func TestRegisterCreatesUserWithPasswordHash(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewService(repository, NewTokenManager("test-secret", 604800))

	user, err := service.Register(context.Background(), RegisterRequest{
		Phone:    " 13800138000 ",
		Password: "123456",
		Nickname: " zhangsan ",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if user.ID != 10001 {
		t.Fatalf("unexpected user id: %d", user.ID)
	}
	if repository.createParams.Phone != "13800138000" {
		t.Fatalf("unexpected phone: %s", repository.createParams.Phone)
	}
	if repository.createParams.Nickname != "zhangsan" {
		t.Fatalf("unexpected nickname: %s", repository.createParams.Nickname)
	}
	if repository.createParams.PasswordHash == "123456" {
		t.Fatal("password was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.createParams.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("password hash does not match password: %v", err)
	}
}

func TestRegisterRejectsInvalidRequest(t *testing.T) {
	service := NewService(&fakeUserRepository{}, NewTokenManager("test-secret", 604800))

	_, err := service.Register(context.Background(), RegisterRequest{
		Phone:    "13800138000",
		Password: "12345",
		Nickname: "zhangsan",
	})
	if !errors.Is(err, ErrInvalidRegisterRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestRegisterReturnsDuplicatePhoneError(t *testing.T) {
	service := NewService(&fakeUserRepository{
		createErr: ErrPhoneAlreadyRegistered,
	}, NewTokenManager("test-secret", 604800))

	_, err := service.Register(context.Background(), RegisterRequest{
		Phone:    "13800138000",
		Password: "123456",
		Nickname: "zhangsan",
	})
	if !errors.Is(err, ErrPhoneAlreadyRegistered) {
		t.Fatalf("expected duplicate phone error, got %v", err)
	}
}

func TestLoginReturnsToken(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := NewService(&fakeUserRepository{
		findUser: model.User{
			ID:           10001,
			Phone:        "13800138000",
			PasswordHash: string(passwordHash),
			Nickname:     "zhangsan",
		},
	}, NewTokenManager("test-secret", 604800))

	response, err := service.Login(context.Background(), LoginRequest{
		Phone:    "13800138000",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if response.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	if response.ExpiresIn != 604800 {
		t.Fatalf("unexpected expires_in: %d", response.ExpiresIn)
	}
	if response.User.ID != 10001 {
		t.Fatalf("unexpected user id: %d", response.User.ID)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := NewService(&fakeUserRepository{
		findUser: model.User{
			ID:           10001,
			Phone:        "13800138000",
			PasswordHash: string(passwordHash),
			Nickname:     "zhangsan",
		},
	}, NewTokenManager("test-secret", 604800))

	_, err = service.Login(context.Background(), LoginRequest{
		Phone:    "13800138000",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidPhoneOrPassword) {
		t.Fatalf("expected invalid phone or password error, got %v", err)
	}
}

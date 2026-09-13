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

func TestRegisterCreatesUserWithPasswordHash(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewService(repository)

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
	service := NewService(&fakeUserRepository{})

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
	})

	_, err := service.Register(context.Background(), RegisterRequest{
		Phone:    "13800138000",
		Password: "123456",
		Nickname: "zhangsan",
	})
	if !errors.Is(err, ErrPhoneAlreadyRegistered) {
		t.Fatalf("expected duplicate phone error, got %v", err)
	}
}

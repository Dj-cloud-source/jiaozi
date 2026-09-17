package admin

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/auth"
	"jiaozi/internal/model"
)

type fakeRepository struct {
	admin model.Admin
	err   error
}

func (r *fakeRepository) FindByUsername(ctx context.Context, username string) (model.Admin, error) {
	if r.err != nil {
		return model.Admin{}, r.err
	}

	return r.admin, nil
}

func TestLoginReturnsAdminToken(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tokens := auth.NewTokenManager("test-secret", 604800)
	service := NewService(&fakeRepository{
		admin: model.Admin{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(passwordHash),
		},
	}, tokens)

	response, err := service.Login(context.Background(), LoginRequest{
		Username: " admin ",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	adminID, err := tokens.ParseAdmin(response.AccessToken)
	if err != nil {
		t.Fatalf("parse admin token failed: %v", err)
	}
	if adminID != 1 {
		t.Fatalf("unexpected admin id: %d", adminID)
	}
	if response.Admin.Username != "admin" {
		t.Fatalf("unexpected username: %s", response.Admin.Username)
	}
}

func TestLoginRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}, auth.NewTokenManager("test-secret", 604800)).Login(context.Background(), LoginRequest{
		Username: "",
		Password: "123456",
	})
	if !errors.Is(err, ErrInvalidLoginRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	_, err = NewService(&fakeRepository{
		admin: model.Admin{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(passwordHash),
		},
	}, auth.NewTokenManager("test-secret", 604800)).Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidUsernameOrPassword) {
		t.Fatalf("expected invalid username or password error, got %v", err)
	}
}

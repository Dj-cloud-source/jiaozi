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
	createParams CreateAdminParams
	admin model.Admin
	err   error
}

func (r *fakeRepository) Create(ctx context.Context, params CreateAdminParams) (model.Admin, error) {
	r.createParams = params
	if r.err != nil {
		return model.Admin{}, r.err
	}

	return model.Admin{
		ID:           1,
		Username:     params.Username,
		PasswordHash: params.PasswordHash,
	}, nil
}

func (r *fakeRepository) FindByUsername(ctx context.Context, username string) (model.Admin, error) {
	if r.err != nil {
		return model.Admin{}, r.err
	}

	return r.admin, nil
}

func TestCreateCreatesAdminWithPasswordHash(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, auth.NewTokenManager("test-secret", 604800))

	admin, err := service.Create(context.Background(), CreateAdminRequest{
		Username: " admin ",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("create admin failed: %v", err)
	}

	if admin.Username != "admin" {
		t.Fatalf("unexpected username: %s", admin.Username)
	}
	if repository.createParams.PasswordHash == "123456" {
		t.Fatal("password was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.createParams.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("password hash does not match password: %v", err)
	}
}

func TestCreateRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}, auth.NewTokenManager("test-secret", 604800)).Create(context.Background(), CreateAdminRequest{
		Username: "admin",
		Password: "12345",
	})
	if !errors.Is(err, ErrInvalidCreateAdminRequest) {
		t.Fatalf("expected invalid create admin request error, got %v", err)
	}
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

package auth

import (
	"errors"
	"testing"
)

func TestTokenSubjectIsolation(t *testing.T) {
	tokens := NewTokenManager("test-secret", 604800)

	userToken, err := tokens.IssueUser(10001)
	if err != nil {
		t.Fatalf("issue user token failed: %v", err)
	}
	adminToken, err := tokens.IssueAdmin(1)
	if err != nil {
		t.Fatalf("issue admin token failed: %v", err)
	}

	userID, err := tokens.ParseUser(userToken)
	if err != nil {
		t.Fatalf("parse user token failed: %v", err)
	}
	if userID != 10001 {
		t.Fatalf("unexpected user id: %d", userID)
	}

	adminID, err := tokens.ParseAdmin(adminToken)
	if err != nil {
		t.Fatalf("parse admin token failed: %v", err)
	}
	if adminID != 1 {
		t.Fatalf("unexpected admin id: %d", adminID)
	}

	if _, err := tokens.ParseAdmin(userToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected user token to be rejected by admin parser, got %v", err)
	}
	if _, err := tokens.ParseUser(adminToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected admin token to be rejected by user parser, got %v", err)
	}
}

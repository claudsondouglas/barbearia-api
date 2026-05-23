package authactions

import (
	"testing"
	"time"
)

func TestVerify_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := generateToken(42, "user@test.com", "user", "access", 15*time.Minute)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	claims, err := Verify(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
	if claims.Email != "user@test.com" {
		t.Errorf("expected email user@test.com, got %s", claims.Email)
	}
}

func TestVerify_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Verify("not.a.valid.token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerify_RefreshTokenRejected(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, _ := generateToken(1, "user@test.com", "user", "refresh", 7*24*time.Hour)
	_, err := Verify(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for refresh token, got %v", err)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, _ := generateToken(1, "user@test.com", "user", "access", -time.Minute)
	_, err := Verify(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

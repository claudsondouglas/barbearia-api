package authactions

import (
	"testing"
	"time"
)

func TestRefresh_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, _ := generateToken(42, "11999999999", "user", "refresh", 7*24*time.Hour)
	tokens, err := Refresh(refreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken != refreshToken {
		t.Error("refresh token should be unchanged")
	}

	claims, err := Verify(tokens.AccessToken)
	if err != nil {
		t.Fatalf("new access token failed verification: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Refresh("not.a.valid.token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestRefresh_AccessTokenRejected(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	accessToken, _ := generateToken(1, "11999999999", "user", "access", 15*time.Minute)
	_, err := Refresh(accessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for access token, got %v", err)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, _ := generateToken(1, "11999999999", "user", "refresh", -time.Minute)
	_, err := Refresh(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for expired refresh token, got %v", err)
	}
}

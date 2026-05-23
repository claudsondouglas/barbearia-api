package authactions

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestLogout_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	err := Logout(context.Background(), rdb, "invalid.token.here")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestLogout_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token, _ := generateToken(1, "user@test.com", "user", "access", -time.Minute)
	err := Logout(context.Background(), rdb, token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestLogout_Success_BlacklistsToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token, _ := generateToken(1, "user@test.com", "user", "access", 15*time.Minute)
	claims, _ := Verify(token)

	if err := Logout(context.Background(), rdb, token); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !IsBlacklisted(context.Background(), rdb, claims.ID) {
		t.Error("expected token to be blacklisted after logout")
	}
}

func TestLogout_RefreshTokenRejected(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token, _ := generateToken(1, "user@test.com", "user", "refresh", 7*24*time.Hour)
	err := Logout(context.Background(), rdb, token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for refresh token, got %v", err)
	}
}

func TestIsBlacklisted_UnknownJTI(t *testing.T) {
	rdb := newTestRedis(t)

	if IsBlacklisted(context.Background(), rdb, "unknown-jti") {
		t.Error("expected false for unknown jti")
	}
}

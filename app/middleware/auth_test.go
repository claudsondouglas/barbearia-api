package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	authactions "barbearia-api/app/auth/actions"
	"barbearia-api/app/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func makeToken(t *testing.T, userID uint, email, role, tokenType string, duration time.Duration) string {
	t.Helper()
	secret := []byte(os.Getenv("JWT_SECRET"))
	claims := authactions.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	return s
}

func setupAuthRouter(rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.GET("/test", middleware.Auth(rdb), func(c *gin.Context) {
		u, _ := middleware.GetAuthUser(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": u.ID,
			"role":    u.Role,
		})
	})
	return r
}

func TestAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rdb := newTestRedis(t)

	r := setupAuthRouter(rdb)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rdb := newTestRedis(t)

	r := setupAuthRouter(rdb)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Token abc123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	r := setupAuthRouter(rdb)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token := makeToken(t, 7, "user@test.com", "user", "access", 15*time.Minute)

	r := setupAuthRouter(rdb)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuth_BlacklistedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token := makeToken(t, 1, "user@test.com", "user", "access", 15*time.Minute)
	_ = authactions.Logout(context.Background(), rdb, token)

	r := setupAuthRouter(rdb)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for blacklisted token, got %d", w.Code)
	}
}

func TestRequireAdmin_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token := makeToken(t, 1, "user@test.com", "user", "access", 15*time.Minute)

	r := gin.New()
	r.GET("/admin", middleware.Auth(rdb), middleware.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAdmin_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	rdb := newTestRedis(t)

	token := makeToken(t, 1, "admin@test.com", "admin", "access", 15*time.Minute)

	r := gin.New()
	r.GET("/admin", middleware.Auth(rdb), middleware.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

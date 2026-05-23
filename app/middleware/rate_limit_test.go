package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"barbearia-api/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.NewStore()
	l := limiter.New(store, limiter.Rate{Period: time.Minute, Limit: 3})

	r := gin.New()
	r.POST("/test", middleware.RateLimit(l), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := range 3 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.NewStore()
	l := limiter.New(store, limiter.Rate{Period: time.Minute, Limit: 3})

	r := gin.New()
	r.POST("/test", middleware.RateLimit(l), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for range 3 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		r.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimit_IsolatesPerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.NewStore()
	l := limiter.New(store, limiter.Rate{Period: time.Minute, Limit: 1})

	r := gin.New()
	r.POST("/test", middleware.RateLimit(l), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for _, ip := range []string{"1.2.3.4", "5.6.7.8"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.RemoteAddr = ip + ":1234"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("ip %s: expected 200, got %d", ip, w.Code)
		}
	}
}

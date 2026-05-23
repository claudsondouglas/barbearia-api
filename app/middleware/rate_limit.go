// Package middleware fornece middlewares reutilizáveis para o servidor Gin.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
)

func RateLimit(l *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := l.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if ctx.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

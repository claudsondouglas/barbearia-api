// Package middleware fornece middlewares reutilizáveis para o servidor Gin.
package middleware

import (
	"net/http"
	"strings"

	authactions "barbearia-api/app/auth/actions"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// AuthUser contém os dados do usuário autenticado populados pelo middleware Auth.
type AuthUser struct {
	ID    uint
	Email string
	Role  string
}

// GetAuthUser retorna o AuthUser armazenado no contexto pelo middleware Auth.
func GetAuthUser(c *gin.Context) (AuthUser, bool) {
	v, ok := c.Get("auth_user")
	if !ok {
		return AuthUser{}, false
	}
	u, ok := v.(AuthUser)
	return u, ok
}

// Auth valida o Bearer token do header Authorization e verifica a blacklist no Redis.
// Popula auth_user no contexto do Gin.
func Auth(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authactions.Verify(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if authactions.IsBlacklisted(c.Request.Context(), rdb, claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		c.Set("auth_user", AuthUser{ID: claims.UserID, Email: claims.Email, Role: claims.Role})
		c.Next()
	}
}

// RequireAdmin rejeita requisições de usuários que não tenham role "admin".
// Deve ser usado após o middleware Auth.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := GetAuthUser(c)
		if !ok || u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

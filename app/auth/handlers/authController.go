// Package authhttp implementa os handlers HTTP para as rotas de autenticação.
package authhttp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"barbearia-api/app"
	authactions "barbearia-api/app/auth/actions"
	"barbearia-api/app/notifications/mail"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Handler implementa os endpoints de autenticação usando o handler base da aplicação.
type Handler struct {
	*app.Handler
	Redis *redis.Client
}

// RefreshInput contém o refresh token para renovação do access token.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordInput contém o e-mail para envio do código OTP.
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

// Login autentica pelo e-mail e senha, retornando access e refresh tokens.
// Responde 401 para credenciais inválidas.
func (h *Handler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := authactions.Login(h.DB, input)
	if err != nil {
		if errors.Is(err, authactions.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Verify valida o Bearer token do header Authorization e retorna os dados do usuário autenticado.
// Responde 401 se o token estiver ausente, inválido, expirado ou revogado.
func (h *Handler) Verify(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := authactions.Verify(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	if authactions.IsBlacklisted(c.Request.Context(), h.Redis, claims.ID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}

// Refresh emite um novo access token a partir de um refresh token válido.
// Responde 401 se o refresh token for inválido ou expirado.
func (h *Handler) Refresh(c *gin.Context) {
	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := authactions.Refresh(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// ForgotPassword envia um OTP por e-mail para o endereço informado.
// Sempre responde 200 independente de o e-mail estar cadastrado,
// para não revelar quais endereços estão registrados.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := mail.NewClient()
	if err != nil {
		log.Printf("forgot password: mail client: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "se o e-mail estiver cadastrado, você receberá um código"})
		return
	}

	if err := authactions.ForgotPassword(h.DB, input.Email, client); err != nil {
		log.Printf("forgot password error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "se o e-mail estiver cadastrado, você receberá um código"})
}

// Register cria um novo usuário e retorna access e refresh tokens.
// Responde 409 se o e-mail já estiver cadastrado.
func (h *Handler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := authactions.Register(h.DB, input)
	if err != nil {
		if errors.Is(err, authactions.ErrEmailAlreadyInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tokens)
}

// Logout invalida o access token adicionando-o à blacklist no Redis.
// Responde 401 se o token estiver ausente ou inválido.
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if err := authactions.Logout(c.Request.Context(), h.Redis, tokenString); err != nil {
		if errors.Is(err, authactions.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// ResetPassword troca a senha do usuário após validar o OTP recebido por e-mail.
// Responde 400 se o código for inválido ou expirado.
func (h *Handler) ResetPassword(c *gin.Context) {
	var input authactions.ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := authactions.ResetPassword(h.DB, input); err != nil {
		if errors.Is(err, authactions.ErrInvalidOrExpiredOTP) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "código inválido ou expirado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "senha atualizada com sucesso"})
}

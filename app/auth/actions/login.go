// Package authactions implementa a lógica de autenticação: login, verificação
// e renovação de tokens JWT, além do fluxo de redefinição de senha via OTP.
package authactions

import (
	"errors"
	"os"
	"time"

	"barbearia-api/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrInvalidCredentials é retornado quando o e-mail não existe ou a senha está incorreta.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Claims contém os dados codificados no token JWT: ID do usuário, e-mail, role e tipo do token.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// Login autentica um usuário pelo e-mail e senha.
// Retorna um par de tokens JWT (access + refresh) em caso de sucesso,
// ou ErrInvalidCredentials se as credenciais forem inválidas.
func Login(db *gorm.DB, input models.LoginInput) (*models.TokenResponse, error) {
	var user models.User
	if result := db.Where("email = ?", input.Email).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, result.Error
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := generateToken(user.ID, user.Email, user.Role, "access", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(user.ID, user.Email, user.Role, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateToken(userID uint, email, role, tokenType string, duration time.Duration) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	claims := Claims{
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
	return token.SignedString(secret)
}

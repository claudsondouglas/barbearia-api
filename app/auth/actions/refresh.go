package authactions

import (
	"os"
	"time"

	"barbearia-api/models"

	"github.com/golang-jwt/jwt/v5"
)

// Refresh valida um refresh token e emite um novo access token com validade de 15 minutos.
// O refresh token original é retornado inalterado.
// Retorna ErrInvalidToken se o token for inválido, expirado ou não for do tipo "refresh".
func Refresh(refreshToken string) (*models.TokenResponse, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	newAccessToken, err := generateToken(claims.UserID, claims.Email, claims.Role, "access", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
	}, nil
}

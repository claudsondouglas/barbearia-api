package authactions

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken é retornado quando o token JWT é inválido, expirado ou não é do tipo "access".
var ErrInvalidToken = errors.New("invalid token")

// Verify valida um access token JWT e retorna as claims contidas nele.
// Rejeita tokens do tipo "refresh" mesmo que a assinatura seja válida.
func Verify(tokenString string) (*Claims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Type != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

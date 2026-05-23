package authactions

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistPrefix = "token:blacklist:"

// Logout invalida o access token adicionando seu jti ao Redis com TTL igual ao tempo restante de expiração.
// Retorna ErrInvalidToken se o token for inválido, expirado ou não for do tipo "access".
func Logout(ctx context.Context, rdb *redis.Client, tokenString string) error {
	claims, err := Verify(tokenString)
	if err != nil {
		return err
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	key := blacklistPrefix + claims.ID
	if err := rdb.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted retorna true se o jti estiver na blacklist do Redis.
func IsBlacklisted(ctx context.Context, rdb *redis.Client, jti string) bool {
	exists, err := rdb.Exists(ctx, blacklistPrefix+jti).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

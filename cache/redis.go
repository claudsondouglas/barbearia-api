// Package cache gerencia a conexão com o Redis.
package cache

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func Connect() *redis.Client {
	opt, err := redis.ParseURL(os.Getenv("REDIS"))
	if err != nil {
		panic("invalid REDIS url: " + err.Error())
	}
	return redis.NewClient(opt)
}

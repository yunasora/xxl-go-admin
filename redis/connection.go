package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect(addr, password string) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Printf("[Redis] 链接成功: %s", addr)
	return nil
}

package db

import (
	"context"
	"net"
	"trisend/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB() (*redis.Client, error) {
	address := net.JoinHostPort(config.DB_HOST, config.DB_PORT)
	opts := &redis.Options{
		Addr:     address,
		Password: config.DB_PASSWORD,
		DB:       config.DB_NAME,
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

package db

import (
	"net"
	"trisend/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB() *redis.Client {
	address := net.JoinHostPort(config.DB_HOST, config.DB_PORT)
	opts := &redis.Options{
		Addr:     address,
		Password: config.DB_PASSWORD,
		DB:       config.DB_NAME,
	}

	return redis.NewClient(opts)
}

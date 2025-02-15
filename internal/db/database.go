package db

import (
	"context"
	"crypto/tls"
	"net"
	"trisend/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB() (*redis.Client, error) {
	var tlsConfig *tls.Config
	address := net.JoinHostPort(config.DB_HOST, config.DB_PORT)

	if config.IsAppEnvProd() {
		tlsConfig = &tls.Config{}
	}

	opts := &redis.Options{
		Addr:     address,
		Password: config.DB_PASSWORD,
		DB:       config.DB_NAME,
		TLSConfig: tlsConfig,
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

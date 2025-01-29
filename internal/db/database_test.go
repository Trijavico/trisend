package db

import (
	"context"
	"testing"
	"time"
	"trisend/config"

	"github.com/testcontainers/testcontainers-go/modules/redis"
)

type Terminate = func(context.Context) error

func NewRedisContainer() (Terminate, error) {
	container, err := redis.Run(
		context.Background(),
		"docker.io/redis:7.2.4",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		return nil, err
	}

	dbHost, err := container.Host(context.Background())
	if err != nil {
		return container.Terminate, err
	}

	dbPort, err := container.MappedPort(context.Background(), "6379/tcp")
	if err != nil {
		return container.Terminate, err
	}

	config.DB_HOST = dbHost
	config.DB_PORT = dbPort.Port()
	config.DB_NAME = 0

	return container.Terminate, err
}

func TestConn(t *testing.T) {
	terminateDB, err := NewRedisContainer()
	defer terminateDB(context.Background())

	if err != nil {
		t.Fatalf("could not start redis container: %v", err)
	}

	redisDB := NewRedisDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisDB.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("redis is down: %v", err)
	}
}

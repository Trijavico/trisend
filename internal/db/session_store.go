package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionStore interface {
	CreateTransitSess(context.Context, string, string, int) error
	GetTransitSessByID(context.Context, string) (string, error)
}

type sessionRedisStore struct {
	db *redis.Client
}

func NewRedisSessionStore(db *redis.Client) SessionStore {
	return &sessionRedisStore{
		db: db,
	}
}

func (store *sessionRedisStore) CreateTransitSess(ctx context.Context, id, code string, expiry int) error {
	key := fmt.Sprintf("login:%s", id)
	err := store.db.Set(ctx, key, code, time.Minute*time.Duration(expiry)).Err()
	if err != nil {
		return fmt.Errorf("failed to create transit session: %s", err)
	}

	return nil
}

func (store *sessionRedisStore) GetTransitSessByID(ctx context.Context, key string) (string, error) {
	key = fmt.Sprintf("login:%s", key)
	value, err := store.db.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return value, nil
}

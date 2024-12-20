package db

import (
	"context"
	"trisend/types"

	"github.com/redis/go-redis/v9"
)

type UserStore interface {
	CreateUser(context.Context, types.CreateUser) error
}

type userRedisStore struct {
	db *redis.Client
}

func NewUserRedisStore(db *redis.Client) UserStore {
	return &userRedisStore{
		db: db,
	}
}

func (store *userRedisStore) CreateUser(ctx context.Context, user types.CreateUser) error {
	return nil
}

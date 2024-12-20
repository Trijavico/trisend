package db

import (
	"context"
	"fmt"
	"trisend/types"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserStore interface {
	CreateUser(context.Context, types.CreateUser) (*types.Session, error)
	UpdateUser(context.Context, types.CreateUser) (*types.Session, error)
	DeleteUser(context.Context, string) error

	AddSSHKey(ctx context.Context, userID string, sshKey string) error
	DeleteSSHKey(ctx context.Context, userID string, sshKey string) error
	GetSSHKeys(ctx context.Context, userID string) ([]string, error)
}

type redisStore struct {
	db *redis.Client
}

func NewUserRedisStore(db *redis.Client) UserStore {
	return &redisStore{
		db: db,
	}
}

func (store *redisStore) CreateUser(ctx context.Context, user types.CreateUser) (*types.Session, error) {
	userID := uuid.NewString()
	key := fmt.Sprintf("user:%s", userID)

	data := map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
		"pfp":      user.Pfp,
	}

	err := store.db.HSet(ctx, key, data).Err()
	if err != nil {
		return nil, err
	}

	createdUser := &types.Session{
		ID:       userID,
		Email:    user.Email,
		Username: user.Username,
		Pfp:      user.Pfp,
	}

	return createdUser, nil
}

func (store *redisStore) UpdateUser(ctx context.Context, user types.CreateUser) (*types.Session, error) {
	userID := uuid.NewString()
	key := fmt.Sprintf("user:%s", userID)

	data := map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
		"pfp":      user.Pfp,
	}

	err := store.db.HSet(ctx, key, data).Err()
	if err != nil {
		return nil, err
	}

	createdUser := &types.Session{
		ID:       userID,
		Email:    user.Email,
		Username: user.Username,
		Pfp:      user.Pfp,
	}

	return createdUser, nil
}

func (store *redisStore) DeleteUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)

	err := store.db.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (store *redisStore) AddSSHKey(ctx context.Context, userID string, sshKey string) error {
	key := fmt.Sprintf("user:%s:ssh_key", userID)

	err := store.db.SAdd(ctx, key, sshKey).Err()
	if err != nil {
		return err
	}

	return nil
}
func (store *redisStore) DeleteSSHKey(ctx context.Context, userID string, sshKey string) error {
	key := fmt.Sprintf("user:%s:ssh_key", userID)

	err := store.db.SRem(ctx, key, sshKey).Err()
	if err != nil {
		return err
	}

	return nil
}

func (store *redisStore) GetSSHKeys(ctx context.Context, userID string) ([]string, error) {
	key := fmt.Sprintf("user:%s:ssh_key", userID)

	sshKeys, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return []string{}, err
	}

	return sshKeys, nil
}
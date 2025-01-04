package db

import (
	"context"
	"fmt"
	"strings"
	"trisend/types"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserStore interface {
	CreateUser(context.Context, types.CreateUser) (*types.Session, error)
	UpdateUser(context.Context, types.CreateUser) (*types.Session, error)
	DeleteUser(context.Context, string) error
	GetByEmail(context.Context, string) (*types.Session, error)

	AddSSHKey(ctx context.Context, userID, title, fingerprint string) error
	DeleteSSHKey(ctx context.Context, userID, sshID string) error
	GetSSHKeys(ctx context.Context, userID string) ([]types.SSHKey, error)
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
	pipe := store.db.TxPipeline()

	data := map[string]string{
		"email":    user.Email,
		"username": user.Username,
		"pfp":      user.Pfp,
	}

	pipe.HSet(ctx, key, data).Err()
	pipe.Set(ctx, fmt.Sprintf("email:%s", user.Email), userID, 0)

	_, err := pipe.Exec(ctx)
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

func (store *redisStore) GetByEmail(ctx context.Context, email string) (*types.Session, error) {
	key := fmt.Sprintf("email:%s", email)

	userKey, err := store.db.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, redis.Nil
	} else if err != nil {
		return nil, err
	}

	userData, err := store.db.HGetAll(ctx, userKey).Result()
	if err != nil {
		return nil, err
	}
	if len(userData) == 0 {
		return nil, redis.Nil
	}

	user := &types.Session{
		ID:       userKey[len("user:"):],
		Email:    userData["email"],
		Username: userData["username"],
		Pfp:      userData["pfp"],
	}

	return user, nil
}

func (store *redisStore) AddSSHKey(ctx context.Context, userID, title, fingerprint string) error {
	sshID := uuid.NewString()

	key := fmt.Sprintf("user:%s:ssh_key", userID)

	pipe := store.db.Pipeline()
	pipe.SAdd(ctx, key, sshID)

	key = "ssh_key:" + sshID
	data := fmt.Sprintf("%s/%s/%s", sshID, title, fingerprint)
	pipe.Set(ctx, key, data, 0)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *redisStore) DeleteSSHKey(ctx context.Context, userID string, sshID string) error {
	pipe := store.db.Pipeline()

	key := fmt.Sprintf("user:%s:ssh_key", userID)
	pipe.SRem(ctx, key, sshID)

	key = fmt.Sprintf("ssh_key:%s", sshID)
	pipe.Del(ctx)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *redisStore) GetSSHKeys(ctx context.Context, userID string) ([]types.SSHKey, error) {
	key := fmt.Sprintf("user:%s:ssh_key", userID)

	sshKeys, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	pipe := store.db.Pipeline()
	keys := make([]types.SSHKey, 0, len(sshKeys))
	cmds := make([]*redis.StringCmd, len(sshKeys))

	for i, data := range sshKeys {
		cmds[i] = pipe.Get(ctx, "ssh_key:"+data)
	}

	_, err = pipe.Exec(ctx)
	for _, cmd := range cmds {
		keyData, err := cmd.Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			return nil, err
		}

		data := strings.Split(keyData, "/")
		keys = append(keys, types.SSHKey{
			ID:          data[0],
			Title:       data[1],
			Fingerprint: data[2],
		})
	}

	return keys, nil
}

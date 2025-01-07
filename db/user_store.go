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
	GetBySSHKey(context.Context, string) (*types.Session, error)

	AddSSHKey(ctx context.Context, userID, title, fingerprint string) error
	DeleteSSHKey(ctx context.Context, sshID string) error
	GetSSHKeys(ctx context.Context, userID string) ([]types.SSHKey, error)
	SSHKeyExists(ctx context.Context, fingerprint string) (bool, error)
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
	pipe.SAdd(ctx, fmt.Sprintf("email:%s", user.Email), userID)

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

	userKey, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	} else if len(userKey) == 0 {
		return nil, redis.Nil
	}

	key = fmt.Sprintf("user:%s", userKey[0])
	userData, err := store.db.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(userData) == 0 {
		return nil, redis.Nil
	}

	user := &types.Session{
		ID:       userKey[0],
		Email:    userData["email"],
		Username: userData["username"],
		Pfp:      userData["pfp"],
	}

	return user, nil
}

func (store *redisStore) GetBySSHKey(ctx context.Context, fingerprint string) (*types.Session, error) {
	key := fmt.Sprintf("ssh_finger:%s:ssh_key", fingerprint)
	data, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, redis.Nil
	}

	userID := strings.Split(data[0], "/")[1]

	key = fmt.Sprintf("user:%s", userID)
	userMap, err := store.db.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	user := &types.Session{
		ID:       userID,
		Email:    userMap["email"],
		Username: userMap["username"],
		Pfp:      userMap["pfp"],
	}

	return user, nil
}

func (store *redisStore) AddSSHKey(ctx context.Context, userID, title, fingerprint string) error {
	sshID := uuid.NewString()

	pipe := store.db.Pipeline()
	// Mapping userID to sshKeyID
	key := fmt.Sprintf("user:%s:ssh_key", userID)
	pipe.SAdd(ctx, key, sshID)

	// Mapping fingerprint to sshKeyID
	key = fmt.Sprintf("ssh_finger:%s:ssh_key", fingerprint)
	data := fmt.Sprintf("%s/%s", sshID, userID)
	pipe.SAdd(ctx, key, data)

	// SSH Keys "Table"
	key = "ssh_key:" + sshID
	data = fmt.Sprintf("%s/%s/%s/%s", sshID, userID, title, fingerprint)
	pipe.SAdd(ctx, key, data)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (store *redisStore) DeleteSSHKey(ctx context.Context, sshID string) error {
	key := fmt.Sprintf("ssh_key:%s", sshID)

	data, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return err
	} else if len(data) == 0 {
		return redis.Nil
	}

	splitted := strings.Split(data[0], "/")
	userID := splitted[1]
	fingerprint := splitted[3]

	pipe := store.db.Pipeline()

	key = fmt.Sprintf("ssh_finger:%s:ssh_key", fingerprint)
	pipe.Del(ctx, key)

	// Delete one from ssh key "Table"
	pipe.Del(ctx, key)

	key = fmt.Sprintf("user:%s:ssh_key", userID)
	pipe.SRem(ctx, key, sshID)

	_, err = pipe.Exec(ctx)
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
	cmds := make([]*redis.StringSliceCmd, len(sshKeys))

	for i, data := range sshKeys {
		cmds[i] = pipe.SMembers(ctx, "ssh_key:"+data)
	}

	_, err = pipe.Exec(ctx)

	for _, cmd := range cmds {
		keyData, err := cmd.Result()
		if err != nil {
			return nil, err
		} else if len(keyData) == 0 {
			continue
		}

		data := strings.Split(keyData[0], "/")
		keys = append(keys, types.SSHKey{
			ID:          data[0],
			Title:       data[2],
			Fingerprint: data[3],
		})
	}

	return keys, nil
}

func (store *redisStore) SSHKeyExists(ctx context.Context, fingerprint string) (bool, error) {
	key := fmt.Sprintf("ssh_finger:%s:ssh_key", fingerprint)

	data, err := store.db.SMembers(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if len(data) == 0 {
		return false, nil
	}

	return true, nil
}

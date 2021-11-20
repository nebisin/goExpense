package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nebisin/goExpense/internal/store"
)

type UserCache struct {
	rdb *redis.Client
}

func NewUserCache(rdb *redis.Client) *UserCache {
	return &UserCache{rdb: rdb}
}

func (c *UserCache) Get(id int64) (*store.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	val, err := c.rdb.Get(ctx, fmt.Sprintf("users.%d", id)).Result()
	switch {
	case err == redis.Nil:
		return nil, ErrRecordNotFound
	case err != nil:
		return nil, err
	case val == "":
		return nil, ErrRecordNotFound
	}

	var user store.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCache) Set(user *store.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return c.rdb.Set(ctx, fmt.Sprintf("users.%d", user.ID), val, 0).Err()
}

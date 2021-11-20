package cache

import (
	"fmt"

	"github.com/go-redis/redis/v8"
)

var (
	ErrRecordNotFound = fmt.Errorf("record not found")
)

type Cache struct {
	RDB  *redis.Client
	User *UserCache
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{
		RDB:  rdb,
		User: NewUserCache(rdb),
	}
}

package cache_test

import (
	"log"
	"os"
	"testing"

	"github.com/nebisin/goExpense/internal/cache"
	"github.com/nebisin/goExpense/pkg/config"
)

var testCache *cache.Cache

func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("../..", "test")
	if err != nil {
		log.Fatal(err)
	}

	rdb, err := cache.ConnectRedis(cfg.RedisConfig.Host, "", cfg.RedisConfig.Port)
	if err != nil {
		log.Fatal(err)
	}

	testCache = cache.NewCache(rdb)

	os.Exit(m.Run())
}

package cache

import (
	"context"
	"encoding/json"
	"kzhikcn/pkg/config"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var cache Cache

type Cache interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

func InitCache(c *config.Config) error {
	var err error
	provider := strings.ToLower(c.Cache.Provider)
	switch provider {
	case "local":
		err = initBadger(c.Cache.Local)
	case "redis":
		err = initRedis(c.Cache.Redis)
	default:
		err = errors.Errorf("unsupported cache provider: %s", c.Storage.Provider)
	}

	return err
}

func initBadger(conf config.CacheLocalConf) error {
	db, err := badger.Open(badger.DefaultOptions(conf.Dir))
	if err != nil {
		return errors.Errorf("failed to open cache: %s", err)
	}

	cache = &BadgerCache{
		db: db,
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			_ = db.RunValueLogGC(0.5)
		}
	}()

	return nil
}

func initRedis(conf config.CacheRedisConf) error {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Username: conf.Username,
		Password: conf.Password.String(),
	})

	cache = &RedisCache{
		client: client,
	}

	return nil
}

func Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return cache.Set(ctx, key, value, ttl)
}

func Get(ctx context.Context, key string) ([]byte, error) {
	return cache.Get(ctx, key)
}

func Delete(ctx context.Context, key string) error {
	return cache.Delete(ctx, key)
}

func Exists(ctx context.Context, key string) (bool, error) {
	return cache.Exists(ctx, key)
}

func SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return Set(ctx, key, []byte(value), ttl)
}

func GetString(ctx context.Context, key string) (string, error) {
	val, err := Get(ctx, key)
	return string(val), err
}

func SetJson(ctx context.Context, key string, val any, ttl time.Duration) error {
	d, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return Set(ctx, key, d, ttl)
}

func GetJson(ctx context.Context, key string, v any) error {
	d, err := Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(d, v)
}

func Keys(v ...string) string {
	return strings.Join(v, ":")
}

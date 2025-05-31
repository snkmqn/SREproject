package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var ctx = context.Background()

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: password,
		DB: db,
	})

	_, err := rdb.Ping(ctx).Result()

	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return &RedisCache{
		client: rdb,
	}
}

func (r *RedisCache) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) InvalidateKeysByPrefix(prefix string) error {
	iter := r.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) Exists(token string) (bool, error) {
	ctx := context.Background()
	res, err := r.client.Exists(ctx, token).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

package cache

import "time"

type CacheService interface {
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
	InvalidateKeysByPrefix (prefix string) error
	Exists(key string) (bool, error)
}

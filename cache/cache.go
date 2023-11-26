package cache

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	CacheByKey(key string, val interface{}, ex time.Duration)
	GetByKey(key string) (string, error)
	GetKeysByPattern(key string, count int64) ([]string, error)
	DeleteKey(key string)
	BatchDeletionKeysByPattern(key string, count int64)
	FlushDB()
	FlushAll()
	Connect(uri string, password string, db int) error
	GetClient() (*CacheClient, error)
	// Ping() error
}

type CacheClient struct {
	IsCluster     bool
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
}
